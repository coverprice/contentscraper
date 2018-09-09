package database

// This set of functions provides an ephemeral in-memory sqlite database used for testing.

import (
	"database/sql"
	"fmt"
)

// Handle for an ephemeral test database
type TestDatabase struct {
	DbConn *sql.DB
}

// Create a new test database and return the handle.
// At the end, explicitly destroy it with $handle.Cleanup()
func NewTestDatabase() (*TestDatabase, error) {
	// ":memory:" is a magic sqlite value that creates an in-memory DB.
	SetConfig(":memory:")

	dbconn, err := NewConnection()
	if err != nil {
		return nil, fmt.Errorf("Could not get new database connection: %v", err)
	}

	return &TestDatabase{
		DbConn: dbconn,
	}, nil
}

func (this *TestDatabase) Cleanup() {
	Shutdown()
}
