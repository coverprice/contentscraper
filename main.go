package main

import (
	"flag"
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/database"
	"github.com/coverprice/contentscraper/drivers"
	"github.com/coverprice/contentscraper/drivers/reddit"
	"github.com/coverprice/contentscraper/server"
	//"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	waitgroup           sync.WaitGroup
	sourceDrivers       []drivers.IDriver
	quitChannels        []chan bool
	harvestInterval     int
	logLevelFlag        string
	isHarvestingEnabled bool
)

func init() {
	flag.IntVar(&harvestInterval, "harvest-interval", 60, "Minutes to wait between harvest runs")
	flag.StringVar(&logLevelFlag, "log-level", "INFO", "One of DEBUG, INFO, WARN, ERROR, FATAL, PANIC")
	flag.BoolVar(&isHarvestingEnabled, "enable-harvest", true, "False to disable harvesting posts")
}

func initialize() (err error) {
	flag.Parse()

	var logLevel log.Level
	if logLevel, err = log.ParseLevel(logLevelFlag); err != nil {
		log.Fatal("Invalid log level: '%s'", logLevel)
	}
	log.SetLevel(logLevel)

	var conf *config.Config

	if conf, err = config.GetConfig(); err != nil {
		return fmt.Errorf("Could not load/parse config file: %v", err)
	}

	log.Debug("Initializing database.")
	database.Initialize(conf.BackendStorePath)

	// Drivers module
	log.Debug("Initializing Drivers module.")
	var sourceLastRunService *drivers.SourceLastRunService
	var dbconn *database.DbConn

	if dbconn, err = database.NewConnection(); err != nil {
		return fmt.Errorf("Could not create new DB connection: %v", err)
	}
	if sourceLastRunService, err = drivers.NewSourceLastRunService(dbconn); err != nil {
		return fmt.Errorf("Could not initialize SourceLastRunService: %v", err)
	}

	// Init RedditDriver
	log.Debug("Initializing Reddit driver.")
	var redditDriver *reddit.RedditDriver
	if dbconn, err = database.NewConnection(); err != nil {
		return fmt.Errorf("Could not create new DB connection: %v", err)
	}
	if redditDriver, err = reddit.NewRedditDriver(dbconn, conf, sourceLastRunService); err != nil {
		return fmt.Errorf("Could not initialize RedditDriver: %v", err)
	}
	sourceDrivers = append(sourceDrivers, redditDriver)

	log.Debug("Initialization complete.")
	return nil
}

func shutdown() {
	log.Debug("Program shutdown initiated")
	for _, ch := range quitChannels {
		ch <- true
	}
	log.Debug("Waiting for goroutines to complete...")
	waitgroup.Wait()
	log.Debug("Closing down database connections...")
	database.Shutdown()
	log.Debug("Shutdown complete")
}

func beginHarvest() {
	var quit_chan = make(chan bool)
	quitChannels = append(quitChannels, quit_chan)

	waitgroup.Add(1)
	go harvestLoop(quit_chan)
}

func harvestLoop(quit chan bool) {
	defer waitgroup.Done()
	for {
		for _, driver := range sourceDrivers {
			log.Debug("Running Harvest...")
			if err := driver.Harvest(); err != nil {
				log.Fatal(err)
				return
			}
		}

		log.Debug("Harvest complete. Waiting for %d minutes...", harvestInterval)
		select {
		case <-time.After(time.Duration(harvestInterval) * time.Minute):
			// Don't do anything, just exit the select{}
		case <-quit:
			return
		}
	}
}

func main() {
	initialize()
	defer shutdown()

	if isHarvestingEnabled {
		beginHarvest()
	}
	server.Launch()
}
