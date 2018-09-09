package database

import (
	"database/sql"
	_ "github.com/mxk/go-sqlite/sqlite3"
)

var connections = make([]*sql.DB, 0)
var dbfilepath string

// Set the configuration for database operations. This only takes one
// parameter: filepath is the directory/filename to store the database.
func SetConfig(filepath string) {
	dbfilepath = filepath
}

// NewConnection creates a new connection to the backend database, whose
// location has already been specified in a prior SetConfig call.
// It tracks all connections, so that when the program shuts down they
// can all be automatically closed.
func NewConnection() (conn *sql.DB, err error) {
	if dbfilepath == "" {
		panic("Database filename is empty")
	}

	if conn, err = sql.Open("sqlite3", dbfilepath); err != nil {
		return
	}
	connections = append(connections, conn)
	return
}

// Shutdown closes all existing connections.
func Shutdown() {
	for _, conn := range connections {
		conn.Close()
	}
	connections = []*sql.DB{}
}
