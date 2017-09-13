package backingstore

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mxk/go-sqlite/sqlite3"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSqlLite3(t *testing.T) {
	dirpath, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("Could not create TempDirectory", err)
	}
	defer os.RemoveAll(dirpath)

	conn, err := GetConnection(filepath.Join(dirpath, "sqlite3.db"))
	if err != nil {
		t.Error("Could not get a connection", err)
	}

	err = conn.ExecSql(`
        CREATE TABLE x
            ( a INTEGER
            , b TEXT
            , c FLOAT
            )
    `)
	if err != nil {
		t.Error("Could not create a table", err)
	}

	err = conn.ExecSql(`INSERT INTO x (a, b, c) VALUES (1, 'Fruit', 1.234)`)
	if err != nil {
		t.Error("Could not insert a row", err)
	}

	var row *sqlite3.RowMap
	row, err = conn.GetFirstRow(`SELECT * FROM x`)
	if err != nil {
		t.Error("Could not retrieve a row", err)
	}
	if row == nil {
		t.Error("Expected a row")
	}
	var colnames = []string{"a", "b", "c"}
	for _, colname := range colnames {
		if _, ok := (*row)[colname]; !ok {
			t.Error(fmt.Sprintf("Expected row to contain column '%s'", colname))
		}
	}
	if _, ok := (*row)["a"].(int64); !ok {
		t.Error("Expected column 'a' to be an int")
	}
	if _, ok := (*row)["b"].(string); !ok {
		t.Error("Expected column 'b' to be a string")
	}
	if _, ok := (*row)["c"].(float64); !ok {
		t.Error("Expected column 'c' to be a float64")
	}

	err = conn.ExecSql(`INSERT INTO x (a, b, c) VALUES (2, 'Machine', 5.678)`)
	if err != nil {
		t.Error("Could not insert another row", err)
	}

	var rows MultiRowResult
	rows, err = conn.GetAllRows(`SELECT * FROM x ORDER BY a`)
	if err != nil {
		t.Error("Could not retrieve all rows", err)
	}

	if len(rows) != 2 {
		t.Error(fmt.Sprintf("Expected 2 rows, but instead got %d", len(rows)))
	}

	if rows[0]["a"] != int64(1) || rows[1]["a"] != int64(2) {
		t.Error("Values of the rows is not what was expected", spew.Sdump(rows))
	}
}
