package database

import (
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	"github.com/mxk/go-sqlite/sqlite3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type Row sqlite3.RowMap
type MultiRowResult []Row

type DbConn struct {
	conn *sqlite3.Conn
}

func (conn *DbConn) Close() (err error) {
	if conn.conn == nil {
		return nil
	}
	err = conn.conn.Close()
	return
}

func (conn *DbConn) ExecSqlWithArgs(sql string, args sqlite3.NamedArgs) (err error) {
	err = conn.conn.Exec(sql, args)
	return
}

func (conn *DbConn) ExecSql(sql string, params ...interface{}) (err error) {
	var args sqlite3.NamedArgs
	var ok bool
	if len(params) == 1 {
		fmt.Println("Only 1 param")
		_, ok := params[0].(sqlite3.NamedArgs)
		if ok {
			args = params[0].(sqlite3.NamedArgs)
		}
	}
	if !ok {
		args = OrderedParamsToArgs(params...)
	}
	err = conn.conn.Exec(sql, args)
	return
}

func (conn *DbConn) GetFirstRow(sql string, params ...interface{}) (row *Row, err error) {
	var stmt *sqlite3.Stmt
	stmt, err = conn.conn.Query(sql, params...)
	if err == io.EOF {
		// No rows
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	sqlrow := make(sqlite3.RowMap)
	if err = stmt.Scan(sqlrow); err != nil {
		return nil, err
	}
	var retrow = Row(sqlrow)
	return &retrow, nil
}

func (conn *DbConn) GetAllRows(sql string, params ...interface{}) (result MultiRowResult, err error) {
	var stmt *sqlite3.Stmt
	// result = make(MultiRowResult, 0)
	stmt, err = conn.conn.Query(sql, params...)
	if err == io.EOF {
		// No rows
		return result, nil
	} else if err != nil {
		return result, err
	}
	for ; err == nil; err = stmt.Next() {
		sqlrow := make(sqlite3.RowMap)
		if err = stmt.Scan(sqlrow); err != nil {
			return
		}
		var retrow = Row(sqlrow)
		result = append(result, retrow)
	}
	return result, nil
}

func OrderedParamsToArgs(params ...interface{}) (args sqlite3.NamedArgs) {
	args = make(sqlite3.NamedArgs)
	var arg_name = 'a'
	for _, val := range params {
		args[fmt.Sprintf("$%c", arg_name)] = val
		if arg_name == 'z' {
			arg_name = 'A'
		} else if arg_name >= 'Z' && arg_name < 'a' {
			panic("Too many arguments to ExecSql, ran out of named identifiers.")
		} else {
			arg_name += 1
		}
	}
	return args
}

type TestDatabase struct {
	DbConn  *DbConn
	dirpath string
}

func NewTestDatabase(t *testing.T) *TestDatabase {
	dirpath, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("Could not create TempDirectory", err)
	}

	Initialize(filepath.Join(dirpath, "sqlite3.db"))

	dbconn, err := NewConnection()
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
	Shutdown()
	os.RemoveAll(this.dirpath)
}
