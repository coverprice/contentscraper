package reddit

import (
	"fmt"
	"github.com/coverprice/contentscraper/drivers"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	scrape "github.com/coverprice/contentscraper/drivers/reddit/scraper"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"time"
)

// bakingTime_s is how long we leave posts to give them time to get a score.
var bakingTime_s = uint64(1 * 24 * 60 * 60)

// --------------------------------------

type Harvester struct {
	scraper              *scrape.Scraper
	persistence          *persist.Persistence
	sourceLastRunService *drivers.SourceLastRunService
	sources              []types.SubredditSourceConfig
}

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
	return nil
}

func (this *Harvester) pullSource(sourceConfig types.SubredditSourceConfig) (err error) {
	var now = uint64(time.Now().Unix())
	lastRun, err := this.sourceLastRunService.GetSourceLastRunFromId(sourceConfig.GetSourceConfigId())
	if err != nil {
		return err
	}

	// Get new posts up until the time we were last run, minus the bakingTime. This means some posts
	// will be re-retrieved and updated with a more recent score.
	var fromTime = lastRun.DateLastRun - bakingTime_s
	var params = scrape.NewParamsForNew(sourceConfig.Subreddit, fromTime, 10000)
	this.scraper.Start(params)
	if err = this.storeResults(); err != nil {
		return
	}

	// In addition to the new + recently baked posts, also get the "top" feed and persist any posts
	// from there. This leverages reddit's own date+score ranking algorithm to ensure that we're
	// storing popular posts.
	params = scrape.NewParamsForHot(sourceConfig.Subreddit, 90)
	this.scraper.Start(params)
	if err = this.storeResults(); err != nil {
		return
	}

	lastRun.DateLastRun = now
	if err := this.sourceLastRunService.UpsertLastRun(lastRun); err != nil {
		return err
	}
	return nil
}

// Called when the scraper has just started. storeResults will pull the RedditPosts
// from the result channel and persist them to the DB.
func (this *Harvester) storeResults() (err error) {
	var timeout = time.NewTimer(2 * time.Minute)
	defer timeout.Stop()
RetrievePostLoop:
	for {
		if !timeout.Stop() {
			<-timeout.C
		}
		timeout.Reset(2 * time.Minute)

		select {
		case <-timeout.C:
			return fmt.Errorf("Scraping timed out")

		case err, ok := (<-this.scraper.ErrChan):
			if ok {
				return fmt.Errorf("An error occurred in the scraper goroutine: %v", err)
			}

		case post, ok := (<-this.scraper.ResultChan):
			if !ok {
				// Nothing else on postChan, channel must be closed.
				break RetrievePostLoop
			}
			if err = this.persistence.StorePost(post); err != nil {
				// Signal to scraper goroutine that it should quit.
				this.scraper.KillChan <- true
				return
			}
		}
	}
	return nil
}
