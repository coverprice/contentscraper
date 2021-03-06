package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/database"
	"github.com/coverprice/contentscraper/drivers"
	"github.com/coverprice/contentscraper/drivers/reddit"
	"github.com/coverprice/contentscraper/server"
	"github.com/coverprice/contentscraper/toolbox"
	//"github.com/davecgh/go-spew/spew"

	"path"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	waitgroup        sync.WaitGroup
	sourceDrivers    []drivers.IDriver
	quitChannels     []chan bool
	harvestInterval  int
	logFilename      string
	logLevelFlag     string
	isHarvestEnabled bool
	webServer        *server.Server
	port             int
)

func init() {
	flag.IntVar(&harvestInterval, "harvest-interval", 60*6, "Minutes to wait between harvest runs")
	flag.StringVar(&logFilename, "logfile", "", "Log to the given file. (absolute or relative to storage directory)")
	flag.BoolVar(&isHarvestEnabled, "enable-harvest", true, "False to disable harvesting posts")
	flag.IntVar(&port, "port", 8080, "Port to listen on")
}

func initialize() (err error) {
	flag.Parse()

	// TODO: a chicken-egg situation exists where the config module defines the storage directory,
	// but logs need to be configured before that. Really, the modules should be totally independent.
	// So to fix this, the logFilename should be used as-is if it's not absolute.
	if logFilename != "" && !path.IsAbs(logFilename) {
		logFilename = filepath.Join(config.StorageDir(), logFilename)
	}
	toolbox.InitLogging(logFilename)

	var conf *config.Config
	if conf, err = config.GetConfig(); err != nil {
		return fmt.Errorf("Could not load/parse config file: %v", err)
	}

	database.SetConfig(conf.BackendStorePath)

	// Init RedditDriver
	log.Debug("Initializing Reddit driver.")
	var dbconn1, dbconn2 *sql.DB
	var redditDriver *reddit.RedditDriver
	if dbconn1, err = database.NewConnection(); err != nil {
		return fmt.Errorf("Could not create DB connection [1]: %v", err)
	}
	if dbconn2, err = database.NewConnection(); err != nil {
		return fmt.Errorf("Could not create DB connection [2]: %v", err)
	}
	if redditDriver, err = reddit.NewRedditDriver(dbconn1, dbconn2, conf); err != nil {
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
	log.Info("Program shutdown initiated")
	for _, ch := range quitChannels {
		ch <- true
	}
	log.Debug("Waiting for goroutines to complete...")
	waitgroup.Wait()
	log.Debug("Closing down database connections...")
	database.Shutdown()
	log.Info("Shutdown complete")
}

func beginHarvest() {
	var quitChan = make(chan bool)
	quitChannels = append(quitChannels, quitChan)

	waitgroup.Add(1)
	go harvestLoop(quitChan)
}

func harvestLoop(quit chan bool) {
	defer waitgroup.Done()

	// The time.NewTimer provides a mechanism for running periodic function calls.
	// When the timer alarm goes off, it sends an arbitrary value to the channel
	// it created when it was initialized. Doing a select{} on this channel allows
	// us to wait for the alarm. When the alarm goes off, it's necessary to stop
	// and reset the timer.
	var timeout = time.NewTimer(time.Duration(harvestInterval) * time.Minute)
	timeout.Stop()
	defer timeout.Stop()

	for {
		for _, driver := range sourceDrivers {
			log.Info("Harvesting...")
			if err := driver.Harvest(); err != nil {
				log.Fatal(err)
				return
			}
		}

		log.Infof("Harvest complete. Waiting for %d minutes...", harvestInterval)
		if !timeout.Stop() && len(timeout.C) > 0 {
			// If we don't check for the channel length, this channel clear blocks forever...
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

	if isHarvestEnabled {
		beginHarvest()
	}

	log.Info("Launching web service")
	go webServer.Launch()
	sig := make(chan os.Signal)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Infof("Received signal %s, shutting down", s)
	os.Exit(0)
}
