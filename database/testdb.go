package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type TestDatabase struct {
	DbConn  *sql.DB
	dirpath string
}

func NewTestDatabase() (*TestDatabase, error) {
	dirpath, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, fmt.Errorf("Could not create TempDirectory: %v", err)
	}

	Initialize(filepath.Join(dirpath, "sqlite3.db"))

	dbconn, err := NewConnection()
	if err != nil {
		os.RemoveAll(dirpath)
		return nil, fmt.Errorf("Could not get new database connection: %v", err)
	}

	return &TestDatabase{
		DbConn:  dbconn,
		dirpath: dirpath,
	}, nil
}

func (this *TestDatabase) Cleanup() {
	Shutdown()
	os.RemoveAll(this.dirpath)
}
