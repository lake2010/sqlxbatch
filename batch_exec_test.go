package sqlxbatch

import (
	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_assert "github.com/stretchr/testify/require"
)

func TestBatchInsertSingle(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchInserter(dbTx,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s",
		3)
	assert.NoError(err)

	b.AddN(nil, "roobs", "dev")

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(1, len(rows))
	assert.Equal(1, rows[0].Id)
	assert.Equal("roobs", rows[0].Name)
	assert.Equal("dev", rows[0].Other)
}

func TestBatchInsertSingleReuse(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchInserter(dbTx,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s",
		3)
	assert.NoError(err)

	b.AddN(nil, "roobs", "dev")

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(1, len(rows))
	assert.Equal(1, rows[0].Id)
	assert.Equal("roobs", rows[0].Name)
	assert.Equal("dev", rows[0].Other)

	b.AddN(nil, "roobs", "dev")

	err = b.BatchExec()
	assert.NoError(err)

	rows = make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(2, len(rows))
	assert.Equal(2, rows[1].Id)
	assert.Equal("roobs", rows[1].Name)
	assert.Equal("dev", rows[1].Other)
}

func TestBatchInsertMultiple(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchInserter(dbTx,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s",
		3)
	assert.NoError(err)

	b.AddN(nil, "roobs", "dev")
	b.AddN(nil, "bb", "boss")

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(2, len(rows))
	assert.Equal(1, rows[0].Id)
	assert.Equal("roobs", rows[0].Name)
	assert.Equal("dev", rows[0].Other)
	assert.Equal(2, rows[1].Id)
	assert.Equal("bb", rows[1].Name)
	assert.Equal("boss", rows[1].Other)
}

func TestBatchInsertMutilpleCustomTpl(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchExecer(dbTx,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s",
		2, "(NULL, ?, ?)")
	assert.NoError(err)

	b.AddN("roobs", "dev")
	b.AddN("bb", "boss")

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(2, len(rows))
	assert.Equal(1, rows[0].Id)
	assert.Equal("roobs", rows[0].Name)
	assert.Equal("dev", rows[0].Other)
	assert.Equal(2, rows[1].Id)
	assert.Equal("bb", rows[1].Name)
	assert.Equal("boss", rows[1].Other)
}

func testBatchUpdatePrep(t *testing.T, assert *_assert.Assertions, dbTx *sqlx.Tx) {
	b, err := NewBatchInserter(dbTx,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s", 3)
	assert.NoError(err)

	b.AddN(nil, "roobs", "dev")
	b.AddN(nil, "bb", "boss")

	err = b.BatchExec()
	assert.NoError(err)
}

func TestBatchUpdate(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	testBatchUpdatePrep(t, assert, dbTx)

	b, err := NewBatchUpdater(dbTx,
		"UPDATE mytable SET other = 'nub'"+
			"WHERE name IN(%s)", 1)
	assert.NoError(err)

	b.AddN("roobs")
	b.AddN("bb")

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(2, len(rows))
	assert.Equal(1, rows[0].Id)
	assert.Equal("roobs", rows[0].Name)
	assert.Equal("nub", rows[0].Other)
	assert.Equal(2, rows[1].Id)
	assert.Equal("bb", rows[1].Name)
	assert.Equal("nub", rows[1].Other)
}

func TestBatchUpdateWithBaseArg(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	testBatchUpdatePrep(t, assert, dbTx)

	b, err := NewBatchUpdater(dbTx,
		"UPDATE mytable SET other = 'nub' "+
			"WHERE id > ? AND name IN(%s) AND other = ?", 1)
	assert.NoError(err)

	b.AddN("roobs")
	b.AddN("bb")

	b.AddBaseArg(0, BASE_ARG_BEFORE)
	b.AddBaseArg("dev", BASE_ARG_AFTER)

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(2, len(rows))
	assert.Equal(1, rows[0].Id)
	assert.Equal("roobs", rows[0].Name)
	assert.Equal("nub", rows[0].Other)
	assert.Equal(2, rows[1].Id)
	assert.Equal("bb", rows[1].Name)
	assert.Equal("boss", rows[1].Other)
}

func TestBaseArgMax(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchUpdater(dbTx,
		"UPDATE mytable SET other = 'nub' "+
			"WHERE id > ? AND name IN(%s) AND other = ?", 1)
	assert.NoError(err)

	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.NoError(err)

	// 1th base arg should err
	err = b.AddBaseArg(0, BASE_ARG_AFTER)
	assert.Error(err)
}

func TestBatchInsertEmpty(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchInserter(dbTx,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s",
		3)
	assert.NoError(err)

	err = b.BatchExec()
	assert.NoError(err)
}

func TestBatchInsertMultipleBatches(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchInserter(dbTx,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s",
		3)
	assert.NoError(err)

	// lower max sql place holders to trigger chunking
	b.maxSqlPlaceHolders = 3

	b.AddN(nil, "roobs", "dev")
	b.AddN(nil, "bb", "boss")

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = dbTx.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(2, len(rows))
	assert.Equal(1, rows[0].Id)
	assert.Equal("roobs", rows[0].Name)
	assert.Equal("dev", rows[0].Other)
	assert.Equal(2, rows[1].Id)
	assert.Equal("bb", rows[1].Name)
	assert.Equal("boss", rows[1].Other)
}

func TestBatchInsertBadQry(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	b, err := NewBatchInserter(dbTx,
		"THIS IS A A BAD QUERY",
		3)
	assert.NoError(err)

	b.AddN(nil, "roobs", "dev")

	err = b.BatchExec()
	assert.Error(err)
}

func TestBatchUpdateUnsupportedCols(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	dbTx, err := db.Beginx()
	assert.NoError(err)
	defer dbTx.Rollback()

	_, err = NewBatchUpdater(dbTx,
		"UPDATE mytable SET other = 'nub'"+
			"WHERE name IN(%s)", 2)
	assert.Error(err)
}

func TestBatchInsertConcurrent(t *testing.T) {
	t.Parallel()
	assert := _assert.New(t)

	db, closeFn := initDB()
	defer closeFn()

	b, err := NewBatchInserter(db,
		"INSERT INTO mytable "+
			"(id, name, other) "+
			"VALUES %s",
		3)
	assert.NoError(err)
	// use 8 workers
	b.UseNWorkers(8)
	// force small batches
	b.maxSqlPlaceHolders = 4

	m := 1000

	expectedRows := make(map[string]string, m)
	for i := 0; i < m; i++ {
		name := fmt.Sprintf("name%d", i)
		other := fmt.Sprintf("other%d", i)

		b.AddN(nil, name, other)
		expectedRows[name] = other
	}

	err = b.BatchExec()
	assert.NoError(err)

	rows := make([]myTableRow, 0)
	err = db.Select(&rows, "SELECT * FROM mytable")
	assert.NoError(err)

	assert.Equal(m, len(rows))

	for _, row := range rows {
		_, expectedRowExists := expectedRows[row.Name]
		assert.True(expectedRowExists)
		assert.Equal(row.Other, expectedRows[row.Name])

		delete(expectedRows, row.Name)
	}
	assert.Equal(0, len(expectedRows))
}
