package reddit

import (
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/coverprice/contentscraper/toolbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func initTestScraper(t *testing.T) *Scraper {
	var conf *config.Config
	var err error

	if conf, err = config.GetConfig(); err != nil {
		t.Errorf("Could not load/parse config file: %v", err)
	}

	var scraper *Scraper
	if scraper, err = NewScraperFromConfig(conf); err != nil {
		t.Errorf("Could not initialize Scraper: %v", err)
	}
	return scraper
}

// Scraper scrapes posts from a given subreddit (does not test pagination
// because only 1 scrape HTTP request is necessary)
func TestCanScrape(t *testing.T) {
	var (
		scraper = initTestScraper(t)
		context = NewContextForHot("bestoflegaladvice")
		err     error
		posts1  []types.RedditPost
		posts2  []types.RedditPost
	)

	context.NumPostsPerScrape = 10

	posts1, err = scraper.GetNextResults(&context)
	require.Nil(t, err, "Non nil error from scraper")

	t.Logf("Retrieved %d posts for 1st request", len(posts1))
	require.NotEqual(t, 0, len(posts1), "Something went wrong, retrieved 0 posts for 1st request")

	t.Logf("Scraper 'after' context is '%s'", context.After)
	posts2, err = scraper.GetNextResults(&context)
	require.Nil(t, err, "Non nil error from scraper")

	t.Logf("Retrieved %d posts for 2nd request", len(posts2))
	require.NotEqual(t, 0, len(posts2), "Something went wrong, retrieved 0 posts for 2nd request")

	// For visual inspection
	t.Logf("------------ FIRST PAGE -------------")
	for _, post := range posts1 {
		t.Logf(toolbox.TruncateStr(post.Title, 40))
	}
	t.Logf("------------ SECOND PAGE -------------")
	for _, post := range posts2 {
		t.Logf(toolbox.TruncateStr(post.Title, 40))
	}
}
