package reddit

import (
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/database"
	"github.com/coverprice/contentscraper/drivers"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	scrape "github.com/coverprice/contentscraper/drivers/reddit/scraper"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
)

// Verify that RedditDriver satisfies the drivers.IDriver interface.
var _ drivers.IDriver = &RedditDriver{}

// Implements drivers.IDriver
type RedditDriver struct {
	harvester *Harvester
	//editor      *Editor
}

func NewRedditDriver(
	dbconn *database.DbConn,
	conf *config.Config,
	sourceLastRunService *drivers.SourceLastRunService,
) (driver *RedditDriver, err error) {
	var scraper *scrape.Scraper
	if scraper, err = scrape.NewScraperFromConfig(conf); err != nil {
		return
	}

	var persistence *persist.Persistence
	if persistence, err = persist.NewPersistence(dbconn); err != nil {
		return
	}

	var harvester *Harvester
	harvester, err = NewHarvester(
		scraper,
		persistence,
		sourceLastRunService,
	)
	if err != nil {
		return
	}

	for _, feed := range conf.Reddit.Feeds {
		for _, subreddit := range feed.Subreddits {
			harvester.AddSourceConfig(types.SubredditSourceConfig{Subreddit: subreddit.Name})
		}
	}

	return &RedditDriver{
		harvester: harvester,
	}, nil
}

func (this *RedditDriver) Harvest() (err error) {
	return this.harvester.Harvest()
}

/*
func (this *RedditDriver) MarkForPublishing() (err error) {
	return err
}
*/

/*
func (this *RedditDriver) GetFeed(feedname drivers.FeedName) (rssfeed drivers.RssFeed, err error) {
	return "", err
}
*/
