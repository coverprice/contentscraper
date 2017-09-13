package main

import (
	"flag"
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/scrapers"
	"github.com/coverprice/contentscraper/scrapers/redditbot"
	"github.com/davecgh/go-spew/spew"
)

var conf config.Config

func initialize() {
	flag.Parse()
	conf, err := config.GetConfigFromFile()
	if err != nil {
		panic(fmt.Errorf("Could not load/parse config file: %v", err))
	}
	if err = scrapers.Initialize(conf); err != nil {
		panic(fmt.Errorf("Could not initialize scrapers: %v", err))
	}
}

func main() {
	initialize()

	reddit_scraper := scrapers.Get("reddit").(redditbot.Scraper)

	posts, err := reddit_scraper.GetPosts("funny")
	if err != nil {
		panic(err)
	}

	for _, post := range posts {
		fmt.Println(spew.Sdump(post))
	}
}
