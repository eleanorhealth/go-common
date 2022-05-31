package hook

import (
	"context"
	"database/sql"
	"testing"

	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func testDB(t *testing.T) *bun.DB {
	assert := assert.New(t)

	dsn := env.Get("POSTGRES_DSN", "")
	if len(dsn) == 0 {
		assert.FailNow("POSTGRES_DSN is empty")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	err := db.ResetModel(context.Background(), (*testModel)(nil))
	assert.NoError(err)

	return db
}

type testModel struct {
	ID   string `bun:",pk"`
	Name string
}

type queryLogger struct {
	queries []string
}

var _ bun.QueryHook = (*queryLogger)(nil)

func (q *queryLogger) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (q *queryLogger) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	q.queries = append(q.queries, event.Query)
}
