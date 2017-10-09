package drivers

import (
	"github.com/coverprice/contentscraper/database"
	"github.com/stretchr/testify/require"
	"testing"
)

type testSourceConfig struct{}

func (this *testSourceConfig) GetSourceConfigId() SourceConfigId {
	return "foobar"
}

const (
	testSourceConfigId = "foobar"
)

func TestCanCreateAndRetrieveSourceConfig(t *testing.T) {
	testDb, err := database.NewTestDatabase()
	if err != nil {
		t.Fatal("Could not init database", err)
	}
	defer testDb.Cleanup()

	sut, err := NewSourceLastRunService(testDb.DbConn)

	// var sourceConfig = testSourceConfig{}

	var slr *SourceLastRun
	slr, err = sut.GetSourceLastRunFromId(testSourceConfigId)

	require.Nil(t, err, "Could not retrieve non-existent SourceLastRun")

	require.NotEqual(t, 0, slr.DateLastRun, "Service should have filled in a default LastRun")
	require.Equal(t, SourceConfigId(testSourceConfigId), slr.SourceConfigId, "Non-matching SourceConfigId")

	t.Logf("Date last run = %d", slr.DateLastRun)

	slr.DateLastRun = 12345
	err = sut.UpsertLastRun(slr)
	require.Nil(t, err, "Could not upsert SourceLastRun")

	slr, err = sut.GetSourceLastRunFromId(testSourceConfigId)
	require.Nil(t, err, "Could not retrieve existing SourceLastRun")
	require.Equal(t, SourceConfigId(testSourceConfigId), slr.SourceConfigId, "")
	require.Equal(t, int64(12345), slr.DateLastRun, "Could not update DateLastRun")
}
