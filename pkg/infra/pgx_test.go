package infra

import (
	"context"
	"testing"

	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

func testPool(t *testing.T) *pgxpool.Pool {
	assert := assert.New(t)

	dsn := env.Get("POSTGRES_DSN", "")
	if len(dsn) == 0 {
		assert.FailNow("POSTGRES_DSN is empty")
	}

	pool, err := PgxPool(context.Background(), dsn, "")
	assert.NotNil(pool)
	assert.NoError(err)

	_, err = pool.Exec(context.Background(), "drop table if exists test")
	assert.NoError(err)

	_, err = pool.Exec(context.Background(), "create table test (id uuid not null, key text not null, value text not null)")
	assert.NoError(err)

	return pool
}

func TestPgxExecutorQuerier_Execute(t *testing.T) {
	assert := assert.New(t)

	pool := testPool(t)
	defer pool.Close()

	executor := NewPgxExecutorQuerier(pool)

	affectedRows, err := executor.Execute(context.Background(), "insert into test (id, key, value) values ($1, $2, $3)", uuid.New().String(), "foo", "bar")
	assert.Equal(int64(1), affectedRows)
	assert.NoError(err)
}

func TestPgxExecutorQuerier_Query(t *testing.T) {
	assert := assert.New(t)

	pool := testPool(t)
	defer pool.Close()

	executor := NewPgxExecutorQuerier(pool)

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

func TestPgxExecutorQuerier_QueryRow(t *testing.T) {
	assert := assert.New(t)

	pool := testPool(t)
	defer pool.Close()

	executor := NewPgxExecutorQuerier(pool)

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
