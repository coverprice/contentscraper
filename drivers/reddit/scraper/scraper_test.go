package reddit

import (
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/toolbox"
	"sync"
	"testing"
	"time"
)

func initTestScraper(t *testing.T) *Scraper {
	var conf *config.Config
	var err error

	if conf, err = config.GetConfigFromFile(); err != nil {
		t.Error("Could not load/parse config file: %v", err)
	}

	var scraper *Scraper
	if scraper, err = NewScraperFromConfig(conf); err != nil {
		t.Error("Could not initialize Scraper: %v", err)
	}
	return scraper
}

// Scraper scrapes posts from a given subreddit (does not test pagination
// because only 1 scrape HTTP request is necessary)
func TestScrapesPosts_SingleScrapeNeeded(t *testing.T) {
	scrapeMaxPosts(10, t)
}

// Scraper scrapes posts from a given subreddit (DOES test pagination
// because > 1 HTTP requests are necessary to get all the posts)
func TestScrapesPosts_MultiScrapeNeeded(t *testing.T) {
	scrapeMaxPosts(200, t)
}

// Note: scraper will get 100 posts at a time, so specifying < 100 posts means just
// 1 scrape will be performed, more than >100 means at least 2 scrapes are performed.
func scrapeMaxPosts(maxPosts int, t *testing.T) {
	var scraper = initTestScraper(t)
	var fromTime = uint64(time.Now().Unix()) - 7*24*60*60

	var scraperParams = NewParamsForNew("funny", fromTime, maxPosts)
	scraper.Start(scraperParams)

	var numPosts int
	var timeout = time.NewTimer(2 * time.Minute)
RetrievePostLoop:
	for {
		if !timeout.Stop() {
			<-timeout.C
		}
		timeout.Reset(30 * time.Second)

		select {
		case <-timeout.C:
			t.Error("Scraping timed out")

		case err, ok := (<-scraper.ErrChan):
			if ok {
				t.Error("An error occurred in the goroutine: %v", err)
			}

		case post, ok := (<-scraper.ResultChan):
			if !ok {
				// Channel is closed.
				break RetrievePostLoop
			}
			t.Logf("Received post: %s", toolbox.TruncateStr(post.Title, 25))
			numPosts++
			if post.IsSticky {
				t.Error("Should have filtered out sticky posts!")
			}
		}
	}
	t.Logf("Retrieved %d posts", numPosts)
	if numPosts == 0 {
		t.Error("Something went wrong, retrieved 0 posts")
	}
}

// Test that killChan will terminate the goroutine.
func TestSendingKillTerminatesScraper(t *testing.T) {
	var fromTime = uint64(time.Now().Unix()) - 7*24*60*60
	var scraper = initTestScraper(t)
	var scraperParams = NewParamsForNew("funny", fromTime, 400)
	scraper.Start(scraperParams)

	var killSent sync.Once
	var timeout = time.NewTimer(2 * time.Minute)
RetrievePostLoop:
	for {
		if !timeout.Stop() {
			<-timeout.C
		}
		timeout.Reset(30 * time.Second)

		select {
		case <-timeout.C:
			t.Error("Scraping timed out")

		case err, ok := (<-scraper.ErrChan):
			if ok {
				t.Error("An error occurred in the goroutine: %v", err)
			} else {
				break RetrievePostLoop
			}

		case post, ok := (<-scraper.ResultChan):
			if !ok {
				// Channel is closed.
				break RetrievePostLoop
			}
			t.Logf("Received post: %s\n", toolbox.TruncateStr(post.Title, 25))
			// As soon as we get a post we know that the scraper is working,
			// so send a message to kill it.
			killSent.Do(func() {
				t.Log("Sending kill signal")
				scraper.KillChan <- true
			})
		}
	}
}
