package reddit

import (
	"database/sql"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers"
	harvest "github.com/coverprice/contentscraper/drivers/reddit/harvester"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	scrape "github.com/coverprice/contentscraper/drivers/reddit/scraper"
	"github.com/coverprice/contentscraper/drivers/reddit/server"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"net/http"
)

// Verify that RedditDriver satisfies the drivers.IDriver interface.
var _ drivers.IDriver = &RedditDriver{}

const (
	baseUrlPath = "/reddit/"
)

// RedditDriver implements drivers.IDriver. It follows the Facade pattern
// and delegates the work of the interface to subordinate classes.
type RedditDriver struct {
	harvester   *harvest.Harvester
	httpHandler *server.HttpHandler
}

func NewRedditDriver(
	harvesterDbconn *sql.DB,
	publisherDbconn *sql.DB,
	conf *config.Config,
	sourceLastRunService *drivers.SourceLastRunService,
) (driver *RedditDriver, err error) {
	var scraper *scrape.Scraper
	if scraper, err = scrape.NewScraperFromConfig(conf); err != nil {
		return
	}

	var persistence *persist.Persistence
	if persistence, err = persist.NewPersistence(harvesterDbconn); err != nil {
		return
	}

	var harvester *harvest.Harvester
	harvester, err = harvest.NewHarvester(
		scraper,
		persistence,
		sourceLastRunService,
	)
	if err != nil {
		return
	}

	htmlViewerRequestHandler := server.NewHtmlViewerRequestHandler(publisherDbconn)
	httpHandler := server.NewHttpHandler(htmlViewerRequestHandler)

	for _, feed := range conf.Reddit.Feeds {
		httpHandler.AddFeed(feed)
		for _, subreddit := range feed.Subreddits {
			harvester.AddSourceConfig(types.SubredditSourceConfig{Subreddit: subreddit.Name})
		}
	}

	return &RedditDriver{
		harvester:   harvester,
		httpHandler: httpHandler,
	}, nil
}

func (this *RedditDriver) GetBaseUrlPath() string {
	return baseUrlPath
}

func (this *RedditDriver) GetFeeds() []drivers.Feed {
	return this.httpHandler.GetFeedsForServer()
}

func (this *RedditDriver) GetHttpHandler() http.Handler {
	return this.httpHandler
}

func (this *RedditDriver) Harvest() (err error) {
	return this.harvester.Harvest()
}
