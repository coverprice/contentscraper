package database

import (
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

type tempRow struct {
	A int
	B string
	C float64
}

func InitTestDb(t *testing.T) *TestDatabase {
	testDb, err := NewTestDatabase()
	if err != nil {
		t.Fatal("Could not init database", err)
	}
	return testDb
}

func createTestTable(t *testing.T, conn *sql.DB) {
	_, err := conn.Exec(`
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

func TestCanConnect(t *testing.T) {
	testdb := InitTestDb(t)
	defer testdb.Cleanup()

	createTestTable(t, testdb.DbConn)
}

func TestCanInsertAndSelect(t *testing.T) {
	testdb := InitTestDb(t)
	defer testdb.Cleanup()

	createTestTable(t, testdb.DbConn)

	var err error
	_, err = testdb.DbConn.Exec(`INSERT INTO x (a, b, c) VALUES (1, 'Fruit', 1.234)`)
	require.Nil(t, err, "Could not insert a row")

	_, err = testdb.DbConn.Exec(`INSERT INTO x (a, b, c) VALUES (2, 'Machine', 5.678)`)
	require.Nil(t, err, "Could not insert a 2nd row")

	_, err = testdb.DbConn.Exec(`INSERT INTO x (a, b, c) VALUES ($a, $b, $c)`, 3, "Whistle", 9.99)
	require.Nil(t, err, "Could not insert a row using parametized queries")

	var rows *sql.Rows
	rows, err = testdb.DbConn.Query(`SELECT a, b, c FROM x WHERE a <= $a ORDER BY a`, 2)
	require.Nil(t, err, "Could not select using parametized queries")
	defer rows.Close()

	var rowCnt int
	for rows.Next() {
		rowCnt++
		var row tempRow
		err = rows.Scan(
			&row.A,
			&row.B,
			&row.C,
		)
		require.Nil(t, err, "Could not Scan row")
	}
	require.Equal(t, 2, rowCnt, "Expected 2 rows")
}

func TestPreparedExec(t *testing.T) {
	testdb := InitTestDb(t)
	defer testdb.Cleanup()
	createTestTable(t, testdb.DbConn)

	var (
		err  error
		stmt *sql.Stmt
		rows *sql.Rows
		cnt  int
	)
	stmt, err = testdb.DbConn.Prepare(`INSERT INTO x (a, b, c) VALUES ($a, $b, $c)`)
	require.Nil(t, err, "Could not prepare a statement")
	require.NotNil(t, stmt, "PreparedStmt was nil")

	_, err = stmt.Exec(1, "Fruit", 1.234)
	require.Nil(t, err, "Exec returned not nil")
	_, err = stmt.Exec(2, "Machine", 3.44)
	require.Nil(t, err, "Exec returned not nil")

	err = testdb.DbConn.QueryRow(`SELECT COUNT(*) AS cnt FROM x`).Scan(&cnt)
	require.Nil(t, err, "Could not select number of rows")
	require.Equal(t, 2, cnt, "Didn't manage to insert 2 posts")

	stmt, err = testdb.DbConn.Prepare(`SELECT a, b, c FROM x WHERE a <= $a ORDER BY a`)
	require.Nil(t, err, "Could not prepare a query statement")
	require.NotNil(t, stmt, "PreparedStmt was nil")

	rows, err = stmt.Query(2)
	require.Nil(t, err, "GetAllRows returned not nil error")
	defer rows.Close()
	var rowCnt int
	var row tempRow
	for rows.Next() {
		rowCnt++
		err = rows.Scan(
			&row.A,
			&row.B,
			&row.C,
		)
		require.Nil(t, err, "Could not Scan row")
	}
	require.Equal(t, 2, rowCnt, "Expected 2 rows")
	require.Equal(t, "Machine", row.B, "Expected value of B to be 'Machine'")
}
