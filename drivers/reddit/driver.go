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

// RedditDriver implements drivers.IDriver. It follows the Facade pattern
// and delegates the work of the interface to subordinate classes.
type RedditDriver struct {
	harvester   *harvest.Harvester
	httpHandler *server.HttpHandler
}

func NewRedditDriver(
	harvesterDbconn *sql.DB,
	viewerDbconn *sql.DB,
	conf *config.Config,
) (driver *RedditDriver, err error) {
	var scraper *scrape.Scraper
	if scraper, err = scrape.NewScraperFromConfig(conf); err != nil {
		return
	}

	var persistenceHarvester *persist.Persistence
	if persistenceHarvester, err = persist.NewPersistence(harvesterDbconn); err != nil {
		return
	}

	var harvester *harvest.Harvester
	harvester, err = harvest.NewHarvester(
		scraper,
		persistenceHarvester,
	)
	if err != nil {
		return
	}

	var persistenceViewer *persist.Persistence
	if persistenceViewer, err = persist.NewPersistence(viewerDbconn); err != nil {
		return
	}
	htmlViewerRequestHandler := server.NewHtmlViewerRequestHandler(persistenceViewer)
	httpHandler := server.NewHttpHandler(htmlViewerRequestHandler)

	for _, feed := range conf.Reddit.Feeds {
		types.FeedRegistry.AddItem(&feed)
	}

	return &RedditDriver{
		harvester:   harvester,
		httpHandler: httpHandler,
	}, nil
}

func (this *RedditDriver) GetBaseUrlPath() string {
	return server.GetBaseUrlPath()
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
