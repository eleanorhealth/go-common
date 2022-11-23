package infra

import (
	"context"
	"database/sql"
	"testing"

	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testDB(t *testing.T) *sql.DB {
	assert := assert.New(t)

	dsn := env.Get("POSTGRES_DSN", "")
	if len(dsn) == 0 {
		assert.FailNow("POSTGRES_DSN is empty")
	}

	db, err := DB(context.Background(), dsn, "")
	assert.NotNil(db)
	assert.NoError(err)

	_, err = db.ExecContext(context.Background(), "drop table if exists test")
	assert.NoError(err)

	_, err = db.ExecContext(context.Background(), "create table test (id uuid not null, key text not null, value text not null)")
	assert.NoError(err)

	return db
}

func TestSQLExecutorQuerier_Execute(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)
	defer db.Close()

	executor := NewSQLExecutorQuerier(db)

	affectedRows, err := executor.Execute(context.Background(), "insert into test (id, key, value) values ($1, $2, $3)", uuid.New().String(), "foo", "bar")
	assert.Equal(int64(1), affectedRows)
	assert.NoError(err)
}

func TestSQLExecutorQuerier_Query(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)
	defer db.Close()

	executor := NewSQLExecutorQuerier(db)

	affectedRows, err := executor.Execute(context.Background(), "insert into test (id, key, value) values ($1, $2, $3)", uuid.New().String(), "foo", "bar")
	assert.Equal(int64(1), affectedRows)
	assert.NoError(err)

	affectedRows, err = executor.Execute(context.Background(), "insert into test (id, key, value) values ($1, $2, $3)", uuid.New().String(), "baz", "cat")
	assert.Equal(int64(1), affectedRows)
	assert.NoError(err)

	affectedRows, err = executor.Execute(context.Background(), "insert into test (id, key, value) values ($1, $2, $3)", uuid.New().String(), "foobar", "bazcat")
	assert.Equal(int64(1), affectedRows)
	assert.NoError(err)

	type record struct {
		ID    string
		Key   string
		Value string
	}

	var records []*record

	err = executor.Query(context.Background(), &records, "select * from test")
	assert.NoError(err)
	assert.Len(records, 3)
}

func TestSQLExecutorQuerier_QueryRow(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)
	defer db.Close()

	executor := NewSQLExecutorQuerier(db)

	affectedRows, err := executor.Execute(context.Background(), "insert into test (id, key, value) values ($1, $2, $3)", uuid.New().String(), "foo", "bar")
	assert.Equal(int64(1), affectedRows)
	assert.NoError(err)

	affectedRows, err = executor.Execute(context.Background(), "insert into test (id, key, value) values ($1, $2, $3)", uuid.New().String(), "baz", "cat")
	assert.Equal(int64(1), affectedRows)
	assert.NoError(err)

	type record struct {
		ID    string
		Key   string
		Value string
	}

	var rec record

	err = executor.QueryRow(context.Background(), &rec, "select * from test where key = $1 limit 1", "foo")
	assert.NoError(err)
	assert.Equal("foo", rec.Key)
	assert.Equal("bar", rec.Value)
}
