package main

import (
	"flag"
	"fmt"
	"github.com/coverprice/contentscraper/backingstore"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/scrapers"
	"github.com/coverprice/contentscraper/scrapers/runner"
	//"github.com/davecgh/go-spew/spew"
)

var conf config.Config

func initialize() {
	flag.Parse()
	conf, err := config.GetConfigFromFile()
	if err != nil {
		panic(fmt.Errorf("Could not load/parse config file: %v", err))
	}

	backingstore.Initialize(conf.BackendStorePath)

	if err = scrapers.Initialize(conf); err != nil {
		panic(fmt.Errorf("Could not initialize scrapers: %v", err))
	}
}

func shutdown() {
	backingstore.Shutdown()
}

func main() {
	initialize()
	defer shutdown()

	reddit_scraper_runner := scrapers.Get("reddit")
	parambag := runner.ParamBag{
		"subreddit": "funny",
	}
	err := reddit_scraper_runner.Run(parambag)
	if err != nil {
		panic(err)
	}
	//fmt.Println(spew.Sdump(post))
}
