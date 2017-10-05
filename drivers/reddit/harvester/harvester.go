package reddit

import (
	"github.com/coverprice/contentscraper/drivers"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	scrape "github.com/coverprice/contentscraper/drivers/reddit/scraper"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	log "github.com/sirupsen/logrus"
	"time"
)

// --------------------------------------

// Harvester controls the process of scraping posts from Reddit sources
// and persisting them.
// It uses a Scraper client to pull posts from a specific source,
// and a Persistence layer to insert/update them. (Posts already stored
// are updated to reflect any changes in their score, deleted status,
// etc).
type Harvester struct {
	scraper              *scrape.Scraper
	persistence          *persist.Persistence
	sourceLastRunService *drivers.SourceLastRunService
	sources              []types.SubredditSourceConfig
	MaxPagesToScrape     int     // Maximum # of pages to scrape (per source)
	MinPostsPerScrape    int     // Min posts in scrape result to continue
	MinNewPostPercent    float64 // Min new posts in scrape result to continue.
}

// Creates a new Harvester instance
func NewHarvester(
	scraper *scrape.Scraper,
	persistence *persist.Persistence,
	sourceLastRunService *drivers.SourceLastRunService,
) (*Harvester, error) {
	return &Harvester{
		scraper:              scraper,
		persistence:          persistence,
		sourceLastRunService: sourceLastRunService,
		sources:              make([]types.SubredditSourceConfig, 0),
		MaxPagesToScrape:     10,
		MinPostsPerScrape:    10,
		MinNewPostPercent:    20.0,
	}, nil
}

func (this *Harvester) AddSourceConfig(sc types.SubredditSourceConfig) {
	this.sources = append(this.sources, sc)
}

func (this *Harvester) Harvest() (err error) {
	for _, source := range this.sources {
		if err = this.pullSource(source); err != nil {
			return
		}
	}
	return
}

func (this *Harvester) pullSource(sourceConfig types.SubredditSourceConfig) (err error) {
	log.Debugf("Pulling from source '%s'", sourceConfig.Subreddit)
	var now = uint64(time.Now().Unix())
	lastRun, err := this.sourceLastRunService.GetSourceLastRunFromId(sourceConfig.GetSourceConfigId())
	if err != nil {
		return err
	}

	var context = scrape.NewContextForHot(sourceConfig.Subreddit)
	numPagesScraped := 0
	for {
		var posts []types.RedditPost
		log.Debugf("Pulling posts from source '%s'", sourceConfig.Subreddit)
		posts, err = this.scraper.GetNextResults(&context)
		if err != nil {
			return
		}
		log.Debugf("Pulled %d posts from source '%s'", len(posts), sourceConfig.Subreddit)
		numPagesScraped++

		numNewPosts := 0
		for _, post := range posts {
			var result persist.StoreResult
			post.TimeStored = now
			if result, err = this.persistence.StorePost(&post); err != nil {
				return
			}
			if result == persist.StoreResult(persist.STORERESULT_NEW) {
				numNewPosts++
			}
		}

		// Decide when to break out of the loop
		if numPagesScraped > this.MaxPagesToScrape {
			// Prevents us from going too far back in time.
			log.Debugf("Breaking out of loop due to max number of pages scraped")
			break
		}
		if len(posts) < this.MinPostsPerScrape {
			// If we have <10 posts in a result, we're probably at the end of
			//  Reddit's available feed.
			log.Debugf("Breaking out of loop because there were %d results but we need a minimum of %d", len(posts), this.MinPostsPerScrape)
			break
		}
		if 100.0*float64(numNewPosts)/float64(len(posts)) < this.MinNewPostPercent {
			// If # of new posts is < certain % of posts scraped in this page,
			// going back further is probably pointless.
			log.Debugf("Breaking out of loop because there were only %d new results out of %d", numNewPosts, len(posts))
			break
		}
	}
	log.Debugf("Completed pulling from source '%s'", sourceConfig.Subreddit)

	lastRun.DateLastRun = now
	if err := this.sourceLastRunService.UpsertLastRun(lastRun); err != nil {
		return err
	}
	return nil
}
