package toolbox

import (
	"github.com/coverprice/contentscraper/database"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type TestDatabase struct {
	DbConn  *database.DbConn
	dirpath string
}

func NewTestDatabase(t *testing.T) *TestDatabase {
	dirpath, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("Could not create TempDirectory", err)
	}

	database.Initialize(filepath.Join(dirpath, "sqlite3.db"))

	dbconn, err := database.NewConnection()
	if err != nil {
		os.RemoveAll(dirpath)
		t.Error("Could not get new database connection", err)
	}

	return &TestDatabase{
		DbConn:  dbconn,
		dirpath: dirpath,
	}
}

func (this *TestDatabase) Cleanup() {
	database.Shutdown()
	os.RemoveAll(this.dirpath)
}
