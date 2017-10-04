package main

import (
	"database/sql"
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
	webServer           *server.Server
	port                int
)

func init() {
	flag.IntVar(&harvestInterval, "harvest-interval", 60, "Minutes to wait between harvest runs")
	flag.StringVar(&logLevelFlag, "log-level", "INFO", "One of DEBUG, INFO, WARN, ERROR, FATAL, PANIC")
	flag.BoolVar(&isHarvestingEnabled, "enable-harvest", true, "False to disable harvesting posts")
	flag.IntVar(&port, "port", 8080, "Port to listen on")
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
	var dbconn1, dbconn2 *sql.DB

	if dbconn1, err = database.NewConnection(); err != nil {
		return fmt.Errorf("Could not create DB connection [1]: %v", err)
	}
	if sourceLastRunService, err = drivers.NewSourceLastRunService(dbconn1); err != nil {
		return fmt.Errorf("Could not initialize SourceLastRunService: %v", err)
	}

	// Init RedditDriver
	log.Debug("Initializing Reddit driver.")
	var redditDriver *reddit.RedditDriver
	if dbconn1, err = database.NewConnection(); err != nil {
		return fmt.Errorf("Could not create DB connection [2]: %v", err)
	}
	if dbconn2, err = database.NewConnection(); err != nil {
		return fmt.Errorf("Could not create DB connection [3]: %v", err)
	}
	if redditDriver, err = reddit.NewRedditDriver(dbconn1, dbconn2, conf, sourceLastRunService); err != nil {
		return fmt.Errorf("Could not initialize RedditDriver: %v", err)
	}
	sourceDrivers = append(sourceDrivers, redditDriver)

	// init web server
	webServer = server.NewServer(port)
	for _, driver := range sourceDrivers {
		webServer.AddDriver(driver)
	}

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
	var timeout = time.NewTimer(time.Duration(harvestInterval) * time.Minute)
	timeout.Stop()
	defer timeout.Stop()

	for {
		for _, driver := range sourceDrivers {
			log.Debug("Running Harvest...")
			if err := driver.Harvest(); err != nil {
				log.Fatal(err)
				return
			}
		}

		log.Debug("Harvest complete. Waiting for %d minutes...", harvestInterval)
		if !timeout.Stop() {
			<-timeout.C
		}
		timeout.Reset(time.Duration(harvestInterval) * time.Minute)
		select {
		case <-timeout.C:
			timeout.Stop()
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
	webServer.Launch()
}
