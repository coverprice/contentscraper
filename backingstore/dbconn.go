package backingstore

import (
	"github.com/mxk/go-sqlite/sqlite3"
)

var connections = make([]*DbConn, 0)
var dbfilepath string

func Initialize(filepath string) {
	dbfilepath = filepath
}

func NewConnection() (conn *DbConn, err error) {
	if dbfilepath == "" {
		panic("Database filename is empty")
	}

	var sqlite_conn *sqlite3.Conn
	if sqlite_conn, err = sqlite3.Open(dbfilepath); err != nil {
		return
	}
	conn = &DbConn{conn: sqlite_conn}
	connections = append(connections, conn)
	return
}

func Shutdown() {
	for _, conn := range connections {
		conn.Close()
	}
}
