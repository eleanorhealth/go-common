package bao2

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/eleanorhealth/go-common/pkg/bao2/hook"
	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/google/uuid"
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

	err := db.ResetModel(context.Background(), db, (*testModel)(nil), (*testRelatedModel)(nil))
	assert.NoError(err)

	return db
}

type testModel struct {
	ID      string `bun:",pk"`
	Name    string
	Related *testRelatedModel `bun:"rel:has-one,join:id=test_model_id"`
}

type testRelatedModel struct {
	ID          string `bun:",pk"`
	TestModelID string
}

type testHook struct {
	before func(ctx context.Context, db bun.IDB, model any) error
	after  func(ctx context.Context, model any)
}

var _ hook.Hook = (*testHook)(nil)

func (t *testHook) Before(ctx context.Context, db bun.IDB, model any) error {
	if t.before != nil {
		return t.before(ctx, db, model)
	}

	return nil
}

func (t *testHook) After(ctx context.Context, model any) {
	if t.after != nil {
		t.after(ctx, model)
	}
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	id := uuid.New().String()

	insertModel := &testModel{
		ID: id,
		Related: &testRelatedModel{
			ID:          uuid.New().String(),
			TestModelID: id,
		},
	}

	err := Create(context.Background(), db, insertModel)
	assert.NoError(err)

	err = Create(context.Background(), db, insertModel.Related)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Relation("Related").Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestCreate_exists(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Create(context.Background(), db, insertModel)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)

	err = Create(context.Background(), db, insertModel)
	pgErr := &pgdriver.Error{}
	assert.ErrorAs(err, pgErr)
	assert.True(pgErr.IntegrityViolation())
}

func TestCreate_hooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	hook := &testHook{}

	var beforeCreateCalled bool
	hook.before = func(ctx context.Context, db bun.IDB, model any) error {
		beforeCreateCalled = true
		return nil
	}

	var afterCreateCalled bool
	hook.after = func(ctx context.Context, model any) {
		afterCreateCalled = true
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Create(context.Background(), db, insertModel, hook)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
	assert.True(beforeCreateCalled)
	assert.True(afterCreateCalled)
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	id := uuid.New().String()

	insertModel := &testModel{
		ID: id,
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	insertModel.Name = "foo"

	err = Update(context.Background(), db, insertModel)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestUpdate_not_exists(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Update(context.Background(), db, insertModel)
	assert.ErrorIs(err, nil)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestUpdate_hooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	hook := &testHook{}

	var beforeUpdateCalled bool
	hook.before = func(ctx context.Context, db bun.IDB, model any) error {
		beforeUpdateCalled = true
		return nil
	}

	var afterUpdateCalled bool
	hook.after = func(ctx context.Context, model any) {
		afterUpdateCalled = true
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Update(context.Background(), db, insertModel, hook)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
	assert.True(beforeUpdateCalled)
	assert.True(afterUpdateCalled)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	id := uuid.New().String()
	insertModel := &testModel{
		ID: id,
	}
	err := Create(context.Background(), db, insertModel)
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Where("id = ?", id).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestDelete_hooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	hook := &testHook{}

	var beforeDeleteCalled bool
	hook.before = func(ctx context.Context, db bun.IDB, model any) error {
		beforeDeleteCalled = true
		return nil
	}

	var afterDeleteCalled bool
	hook.after = func(ctx context.Context, model any) {
		afterDeleteCalled = true
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel, hook)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)

	assert.True(beforeDeleteCalled)
	assert.True(afterDeleteCalled)
}

func TestTrx(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Trx(context.Background(), db, func(ctx context.Context, tx bun.IDB) error {
		_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
		return err
	})
	assert.NoError(err)

	var beginCount int
	var commitCount int
	for _, q := range qLogger.queries {
		if strings.Contains(q, "BEGIN") {
			beginCount++
		}

		if strings.Contains(q, "COMMIT") {
			commitCount++
		}
	}

	assert.Equal(1, beginCount)
	assert.Equal(1, commitCount)
}

func TestTrx_external_tx(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	tx, err := db.Begin()
	assert.NoError(err)

	err = Trx(context.Background(), tx, func(ctx context.Context, tx bun.IDB) error {
		_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
		return err
	})
	assert.NoError(err)

	err = tx.Commit()
	assert.NoError(err)

	var beginCount int
	var commitCount int
	for _, q := range qLogger.queries {
		if strings.Contains(q, "BEGIN") {
			beginCount++
		}

		if strings.Contains(q, "COMMIT") {
			commitCount++
		}
	}

	assert.Equal(1, beginCount)
	assert.Equal(1, commitCount)
}

func TestTrx_nested_transaction(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Trx(context.Background(), db, func(ctx context.Context, tx bun.IDB) error {
		return Trx(context.Background(), tx, func(ctx context.Context, tx bun.IDB) error {
			_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
			return err
		})
	})
	assert.NoError(err)

	var beginCount int
	var commitCount int
	for _, q := range qLogger.queries {
		if strings.Contains(q, "BEGIN") {
			beginCount++
		}

		if strings.Contains(q, "COMMIT") {
			commitCount++
		}
	}

	assert.Equal(1, beginCount)
	assert.Equal(1, commitCount)
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
