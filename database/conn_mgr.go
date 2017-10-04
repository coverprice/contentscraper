package database

import (
	"database/sql"
	_ "github.com/mxk/go-sqlite/sqlite3"
)

var connections = make([]*sql.DB, 0)
var dbfilepath string

func Initialize(filepath string) {
	dbfilepath = filepath
}

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

func Shutdown() {
	for _, conn := range connections {
		conn.Close()
	}
	connections = []*sql.DB{}
}
