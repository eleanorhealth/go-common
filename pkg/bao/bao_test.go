package bao

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

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

	err := db.ResetModel(context.Background(), (*testModel)(nil), (*testRelatedModel)(nil))
	assert.NoError(err)

	return db
}

type testModelEntity struct {
	ID   string
	Name string
}

type testModel struct {
	ID      string `bun:",pk"`
	Name    string
	Related *testRelatedModel `bun:"rel:has-one,join:id=test_model_id"`
}

func (t *testModel) ToEntity() (*testModelEntity, error) {
	return &testModelEntity{
		ID:   t.ID,
		Name: t.Name,
	}, nil
}

func (t *testModel) FromEntity(entity *testModelEntity) error {
	t.ID = entity.ID
	t.Name = entity.Name

	return nil
}

type testRelatedModelEntity struct {
	ID string
}

type testRelatedModel struct {
	ID          string `bun:",pk"`
	TestModelID string
}

func (t *testRelatedModel) ToEntity() (*testRelatedModelEntity, error) {
	return &testRelatedModelEntity{
		ID: t.ID,
	}, nil
}

func (t *testRelatedModel) FromEntity(entity *testRelatedModelEntity) error {
	t.ID = entity.ID

	return nil
}

func TestSelectQuery_non_pointer(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	model := testModel{}
	query, table, err := SelectQuery(context.Background(), db, model)
	assert.Nil(query)
	assert.Nil(table)
	assert.ErrorIs(err, ErrModelNotPointer)
}

func TestSelectQuery_non_struct_slice_pointer(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	model := 1
	query, table, err := SelectQuery(context.Background(), db, &model)
	assert.Nil(query)
	assert.Nil(table)
	assert.ErrorIs(err, ErrModelNotStructSlicePointer)
}

func TestSelectQuery(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	relatedModel := &testRelatedModel{
		ID:          uuid.New().String(),
		TestModelID: insertModel.ID,
	}
	_, err = db.NewInsert().Model(relatedModel).Exec(context.Background())
	assert.NoError(err)

	model := &testModel{}
	query, table, err := SelectQuery(context.Background(), db, model)
	assert.Equal("test_models", table.Name)
	assert.NoError(err)

	err = query.Scan(context.Background())
	assert.NoError(err)

	assert.NotNil(model.Related)
}

func TestSelectForUpdateQuery(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model := &testModel{}
	query, _, err := SelectForUpdateQuery(context.Background(), db, model, false)
	assert.NoError(err)

	err = query.Scan(context.Background())
	assert.NoError(err)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model")
}

func TestSelectForUpdateQuery_skip_locked(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model := &testModel{}
	query, _, err := SelectForUpdateQuery(context.Background(), db, model, true)
	assert.NoError(err)

	err = query.Scan(context.Background())
	assert.NoError(err)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model SKIP LOCKED")
}

func TestFind(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	insertModel2 := &testModel{
		ID: uuid.New().String(),
	}
	_, err = db.NewInsert().Model(insertModel2).Exec(context.Background())
	assert.NoError(err)

	model, err := Find[testModel](context.Background(), db, nil)
	assert.NoError(err)

	assert.Len(model, 2)

	ids := []string{model[0].ID, model[1].ID}
	assert.Contains(ids, insertModel.ID)
	assert.Contains(ids, insertModel2.ID)
}

func TestFind_query(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID:   uuid.New().String(),
		Name: "foo",
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	insertModel2 := &testModel{
		ID:   uuid.New().String(),
		Name: "bar",
	}
	_, err = db.NewInsert().Model(insertModel2).Exec(context.Background())
	assert.NoError(err)

	model, err := Find[testModel](context.Background(), db, func(q *bun.SelectQuery) {
		q.Where("name = ?", "foo")
	})
	assert.NoError(err)

	assert.Len(model, 1)

	assert.Equal(insertModel, model[0])
}

func TestFind_not_found(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	model, err := Find[testModel](context.Background(), db, nil)
	assert.NoError(err)

	assert.Len(model, 0)
}

func TestFindByID(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByID[testModel](context.Background(), db, insertModel.ID)
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestFindByID_not_found(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	_, err := FindByID[testModel](context.Background(), db, "non-existent-id")
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestFindByIDForUpdate(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByIDForUpdate[testModel](context.Background(), db, insertModel.ID, false)
	assert.NoError(err)

	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model")
}

func TestFindByIDForUpdate_skip_locked(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByIDForUpdate[testModel](context.Background(), db, insertModel.ID, true)
	assert.NoError(err)

	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model SKIP LOCKED")
}

func TestSave(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	err := Save(context.Background(), db, insertModel, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestSave_hooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	var beforeSaveCalled bool
	beforeSave := func(ctx context.Context, db bun.IDB, model *testModel) error {
		beforeSaveCalled = true
		return nil
	}

	var afterSaveCalled bool
	afterSave := func(ctx context.Context, model *testModel) {
		afterSaveCalled = true
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	err := Save(context.Background(), db, insertModel, beforeSave, afterSave)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
	assert.True(beforeSaveCalled)
	assert.True(afterSaveCalled)
}

func TestSave_before_hook_error(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	var beforeSaveErr = errors.New("test")
	beforeSave := func(ctx context.Context, db bun.IDB, model *testModel) error {
		return beforeSaveErr
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	err := Save(context.Background(), db, insertModel, beforeSave, nil)
	assert.ErrorIs(err, beforeSaveErr)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestDelete_hooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	var beforeDeleteCalled bool
	beforeDelete := func(ctx context.Context, db bun.IDB, model *testModel) error {
		beforeDeleteCalled = true
		return nil
	}

	var afterDeleteCalled bool
	afterDelete := func(ctx context.Context, model *testModel) {
		afterDeleteCalled = true
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel, beforeDelete, afterDelete)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)

	assert.True(beforeDeleteCalled)
	assert.True(afterDeleteCalled)
}

func TestDelete_before_hook_error(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	beforeDeleteErr := errors.New("test")
	beforeDelete := func(ctx context.Context, db bun.IDB, model *testModel) error {
		return beforeDeleteErr
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel, beforeDelete, nil)
	assert.ErrorIs(err, beforeDeleteErr)
}

func TestTrx(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := trx(context.Background(), db, func(ctx context.Context, tx bun.IDB) error {
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

	err = trx(context.Background(), tx, func(ctx context.Context, tx bun.IDB) error {
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

	err := trx(context.Background(), db, func(ctx context.Context, tx bun.IDB) error {
		return trx(context.Background(), tx, func(ctx context.Context, tx bun.IDB) error {
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
