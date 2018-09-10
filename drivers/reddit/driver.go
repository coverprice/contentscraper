package reddit

// Implements the IDriver interface for the Reddit content source type

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
	harvesterDbconn *sql.DB,	// DB connection used to store harvested content
	viewerDbconn *sql.DB,		// DB connection used to retrieve harvested content
	conf *config.Config,
) (driver *RedditDriver, err error) {
	// Setup harvester
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

	// Setup Feed viewer
	var persistenceViewer *persist.Persistence
	if persistenceViewer, err = persist.NewPersistence(viewerDbconn); err != nil {
		return
	}
	htmlViewerRequestHandler := server.NewHtmlViewerRequestHandler(persistenceViewer)
	httpHandler := server.NewHttpHandler(htmlViewerRequestHandler)

	// Configure Feeds to view
	for _, feed := range conf.Reddit.Feeds {
		types.FeedRegistry.AddItem(&feed)
	}

	return &RedditDriver{
		harvester:   harvester,
		httpHandler: httpHandler,
	}, nil
}

func (this *RedditDriver) GetBaseUrlPath() string {
	return server.BaseUrlPath
}

func (this *RedditDriver) GetFeeds() []drivers.Feed {
	var ret = make([]drivers.Feed, 0)
	for _, feedregistryitem := range types.FeedRegistry.GetAllItems() {
		ret = append(ret, drivers.Feed{
			Name:              feedregistryitem.RedditFeed.Name,
			Description:       feedregistryitem.RedditFeed.Description,
			Status:            feedregistryitem.Status,
			TimeLastHarvested: feedregistryitem.TimeLastHarvested,
		})
	}
	return ret
}

func (this *RedditDriver) GetHttpHandler() http.Handler {
	return this.httpHandler
}

func (this *RedditDriver) Harvest() (err error) {
	return this.harvester.Harvest()
}
