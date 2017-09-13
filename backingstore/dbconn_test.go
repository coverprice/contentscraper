package backingstore

import (
	"os"
	"testing"
)

// Tests whether the automatic closure works
func TestDbConn(t *testing.T) {
	dirpath := initTestDb(t)
	defer os.RemoveAll(dirpath)

	var err error
	_, err = NewConnection()
	if err != nil {
		t.Error("Could not get a connection", err)
	}

	_, err = NewConnection()
	if err != nil {
		t.Error("Could not get a 2nd connection", err)
	}

	Shutdown()
}
