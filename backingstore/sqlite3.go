package backingstore

import (
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	"github.com/mxk/go-sqlite/sqlite3"
	"io"
)

type MultiRowResult []sqlite3.RowMap

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

func (conn *DbConn) GetFirstRow(sql string, params ...interface{}) (row *sqlite3.RowMap, err error) {
	var stmt *sqlite3.Stmt
	stmt, err = conn.conn.Query(sql, params...)
	if err == io.EOF {
		// No rows
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return_row := make(sqlite3.RowMap)
	if err = stmt.Scan(return_row); err != nil {
		return nil, err
	}
	return &return_row, nil
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
		return_row := make(sqlite3.RowMap)
		if err = stmt.Scan(return_row); err != nil {
			return
		}
		result = append(result, return_row)
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
