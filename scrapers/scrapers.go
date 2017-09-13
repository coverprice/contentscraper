package scrapers

import (
	"fmt"
	"github.com/coverprice/contentscraper/backingstore"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/scrapers/redditbot"
	"github.com/coverprice/contentscraper/scrapers/runner"
	// "github.com/davecgh/go-spew/spew"
)

var scraper_runners = make(map[string]*runner.ScraperRunner, 0)

func Initialize(conf *config.Config) (err error) {
	scraper_runners["reddit"], err = newRedditScraperRunner(conf)
	if err != nil {
		return fmt.Errorf("Could not create Reddit ScraperRunner. %v", err)
	}
	return
}

func Get(name string) *runner.ScraperRunner {
	scraper_runner, ok := scraper_runners[name]
	if !ok {
		panic(fmt.Errorf("Unknown scraper '%s'", name))
	}
	return scraper_runner
}

func newRedditScraperRunner(conf *config.Config) (*runner.ScraperRunner, error) {
	var err error
	var scraper redditbot.Scraper
	scraper, err = redditbot.NewScraper(
		conf.Reddit_secrets.ClientId,
		conf.Reddit_secrets.ClientSecret,
		conf.Reddit_secrets.Username,
		conf.Reddit_secrets.Password,
	)
	if err != nil {
		return nil, err
	}

	var dbconn *backingstore.DbConn
	if dbconn, err = backingstore.NewConnection(); err != nil {
		return nil, err
	}

	var datastore redditbot.DataStore
	if datastore, err = redditbot.NewDataStore(dbconn); err != nil {
		return nil, err
	}
	return &runner.ScraperRunner{
		Scraper:   &scraper,
		Datastore: &datastore,
	}, err
}
