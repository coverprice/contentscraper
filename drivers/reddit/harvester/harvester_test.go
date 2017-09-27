package reddit

import (
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/database"
	"github.com/coverprice/contentscraper/drivers"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	scrape "github.com/coverprice/contentscraper/drivers/reddit/scraper"
	"github.com/coverprice/contentscraper/drivers/reddit/types"

	"github.com/stretchr/testify/require"
	"testing"
)

func TestHarvesterRetrievesAndStoresPosts(t *testing.T) {
	var testDb = database.NewTestDatabase(t)
	defer testDb.Cleanup()

	bakingTime_s = uint64(60 * 60) // Set this to just 1 hour.

	// set up harvester
	var harvester = getSut(t, testDb.DbConn)

	// add a source
	harvester.AddSourceConfig(
		types.SubredditSourceConfig{
			Subreddit: "funny",
		},
	)

	// run the harvester
	t.Log("Beginning harvest")
	if err := harvester.Harvest(); err != nil {
		t.Error("Harvest() failed: ", err)
	}
	t.Log("Harvest complete")

	// verify that the database has posts
	row, err := testDb.DbConn.GetFirstRow(
		`SELECT COUNT(*) AS cnt
         FROM redditpost`,
	)
	if err != nil {
		t.Error("Could not retrieve the count of posts")
	}
	var cnt int64
	var ok bool
	if cnt, ok = (*row)["cnt"].(int64); !ok {
		t.Error("Row count could not be converted to an int64")
	}
	require.NotEqual(t, 0, cnt, "Gathered no posts")
	t.Logf("Harvested %d posts", cnt)
}

func getSut(t *testing.T, dbconn *database.DbConn) *Harvester {
	var conf *config.Config
	var err error

	if conf, err = config.GetConfig(); err != nil {
		t.Error("Could not load/parse config file: ", err)
	}

	var scraper *scrape.Scraper
	if scraper, err = scrape.NewScraperFromConfig(conf); err != nil {
		t.Error("Could not initialize scraper: ", err)
	}

	var persistence *persist.Persistence
	if persistence, err = persist.NewPersistence(dbconn); err != nil {
		t.Error("Could not initialize persistence layer: ", err)
	}

	var sourceLastRunService *drivers.SourceLastRunService
	if sourceLastRunService, err = drivers.NewSourceLastRunService(dbconn); err != nil {
		t.Error("Could not initialize SourceLastRunService: ", err)
	}
	// Set this to just 1 hour so that we don't scrape a whole week
	// for the test.
	sourceLastRunService.DefaultLastRunInterval_s = uint64(60 * 60)

	var harvester *Harvester
	harvester, err = NewHarvester(scraper, persistence, sourceLastRunService)
	if err != nil {
		t.Error("Could not initialize Harvester: ", err)
	}
	return harvester
}
