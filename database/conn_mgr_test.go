package database

import (
	"testing"
)

// Tests whether the automatic closure works
func TestConnectionManagerConn(t *testing.T) {
	testDb := InitTestDb(t)
	defer testDb.Cleanup()

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
