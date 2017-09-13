package scrapers

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/scrapers/redditbot"
	// "github.com/davecgh/go-spew/spew"
)

type Scraper interface{}

var scrapers = make(map[string]Scraper, 0)

func Initialize(conf *config.Config) (err error) {
	scrapers["reddit"], err = redditbot.MakeScraper(conf)
	if err != nil {
		return fmt.Errorf("Could not create Reddit bot. %v", err)
	}
	return
}

func Get(name string) Scraper {
	if scraper, ok := scrapers[name]; ok {
		return scraper
	}
	panic(fmt.Errorf("Unknown scraper '%s'", name))
}
