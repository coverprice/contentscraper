package database

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func InitTestDb(t *testing.T) *TestDatabase {
	testDb, err := NewTestDatabase()
	if err != nil {
		t.Fatal("Could not init database", err)
	}
	return testDb
}

func createTestTable(t *testing.T, conn *DbConn) {
	err := conn.ExecSql(`
        CREATE TABLE x
            ( a INTEGER
            , b TEXT
            , c FLOAT
            )
    `)
	if err != nil {
		t.Error("Could not create a table", err)
	}
}

func verifyRowTypes(t *testing.T, conn *DbConn) {
	var row *Row
	var err error
	row, err = conn.GetFirstRow(`SELECT * FROM x`)
	require.Nil(t, err, "Could not retrieve a row")
	require.NotNil(t, row, "Expected a row")

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
}

func TestCanConnect(t *testing.T) {
	testdb := InitTestDb(t)
	defer testdb.Cleanup()

	createTestTable(t, testdb.DbConn)
}

func TestCanInsertAndSelect(t *testing.T) {
	testdb := InitTestDb(t)
	defer testdb.Cleanup()

	createTestTable(t, testdb.DbConn)

	err := testdb.DbConn.ExecSql(`INSERT INTO x (a, b, c) VALUES (1, 'Fruit', 1.234)`)
	if err != nil {
		t.Error("Could not insert a row", err)
	}

	verifyRowTypes(t, testdb.DbConn)

	err = testdb.DbConn.ExecSql(`INSERT INTO x (a, b, c) VALUES (2, 'Machine', 5.678)`)
	if err != nil {
		t.Error("Could not insert another row", err)
	}

	var rows MultiRowResult
	rows, err = testdb.DbConn.GetAllRows(`SELECT * FROM x ORDER BY a`)
	if err != nil {
		t.Error("Could not retrieve all rows", err)
	}

	require.Len(t, rows, 2, "Insufficient rows retrieved")

	if rows[0]["a"] != int64(1) || rows[1]["a"] != int64(2) {
		t.Error("Values of the rows is not what was expected", spew.Sdump(rows))
	}
}

func TestNamedParams(t *testing.T) {
	testdb := InitTestDb(t)
	defer testdb.Cleanup()

	createTestTable(t, testdb.DbConn)

	require.Nil(t,
		testdb.DbConn.ExecSql(`INSERT INTO x (a, b, c) VALUES ($a, $b, $c)`, 1, "Fruit", 1.234),
		"Could not insert a row",
	)

	verifyRowTypes(t, testdb.DbConn)
}
