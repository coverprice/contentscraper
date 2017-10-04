package reddit

import (
	"database/sql"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/database"
	"github.com/coverprice/contentscraper/drivers"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	scrape "github.com/coverprice/contentscraper/drivers/reddit/scraper"
	"github.com/coverprice/contentscraper/drivers/reddit/types"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHarvesterRetrievesAndStoresPosts(t *testing.T) {
	testDb, err := database.NewTestDatabase()
	if err != nil {
		t.Fatal("Could not init database", err)
	}
	defer testDb.Cleanup()

	log.SetLevel(log.DebugLevel)
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
	var cnt int
	err = testDb.DbConn.QueryRow(
		`SELECT COUNT(*)
         FROM redditpost`,
	).Scan(&cnt)
	if err != nil {
		t.Error("Could not retrieve the count of posts")
	}
	require.NotEqual(t, 0, cnt, "Gathered no posts")
	t.Logf("Harvested %d posts", cnt)
}

func getSut(t *testing.T, dbconn *sql.DB) *Harvester {
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

	var harvester *Harvester
	harvester, err = NewHarvester(scraper, persistence, sourceLastRunService)
	if err != nil {
		t.Error("Could not initialize Harvester: ", err)
	}
	// Reduce the defaults so the test completes within a reasonable time limit.
	harvester.MaxPagesToScrape = 2

	return harvester
}
