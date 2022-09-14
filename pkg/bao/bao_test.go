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

	err := db.ResetModel(context.Background(), db, (*testModel)(nil), (*testRelatedModel)(nil))
	assert.NoError(err)

	return db
}

type testModel struct {
	ID      string `bun:",pk"`
	Name    string
	Related *testRelatedModel `bun:"rel:has-one,join:id=test_model_id" bao:",update"`
}

type testRelatedModel struct {
	ID          string `bun:",pk"`
	TestModelID string
}

func TestStore_SelectQuery_non_struct_slice_pointer(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	model := 1

	query, table, err := SelectQuery(context.Background(), db, &model)
	assert.Nil(query)
	assert.Nil(table)
	assert.ErrorIs(err, ErrModelNotStructOrSlice)
}

func TestStore_SelectQuery_struct(t *testing.T) {
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

	query.Relation("Related")

	err = query.Scan(context.Background())
	assert.NoError(err)

	assert.NotNil(model.Related)
}

func TestStore_SelectQuery_slice(t *testing.T) {
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

	var model []*testModel

	query, table, err := SelectQuery(context.Background(), db, &model)
	assert.Equal("test_models", table.Name)
	assert.NoError(err)

	query.Relation("Related")

	err = query.Scan(context.Background())
	assert.NoError(err)

	assert.NotNil(model[0].Related)
}

func TestStore_SelectForUpdateQuery(t *testing.T) {
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
	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model")
}

func TestStore_SelectForUpdateQuery_skip_locked(t *testing.T) {
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
	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model SKIP LOCKED")
}

func TestStore_Find(t *testing.T) {
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

func TestStore_Find_query(t *testing.T) {
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

func TestStore_Find_not_found(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	model, err := Find[testModel](context.Background(), db, nil)
	assert.NoError(err)

	assert.Len(model, 0)
}

func TestStore_FindFirst_query(t *testing.T) {
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

	model, err := FindFirst[testModel](context.Background(), db, func(q *bun.SelectQuery) {
		q.Where("name = ?", "foo")
	})
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestStore_FindFirst_not_found(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	model, err := FindFirst[testModel](context.Background(), db, nil)
	assert.Nil(model)
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestStore_FindByID(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByID[testModel](context.Background(), db, insertModel.ID, nil)
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestStore_FindByID_query(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByID[testModel](context.Background(), db, insertModel.ID, func(q *bun.SelectQuery) {
		q.Where("1 = 1")
	})
	assert.NoError(err)
	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "1 = 1")
}

func TestStore_FindByID_not_found(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	model, err := FindByID[testModel](context.Background(), db, "non-existent-id", nil)
	assert.Nil(model)
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestStore_FindByIDForUpdate(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByIDForUpdate[testModel](context.Background(), db, insertModel.ID, false, nil)
	assert.NoError(err)
	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model")
}

func TestStore_FindByIDForUpdate_query(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByIDForUpdate[testModel](context.Background(), db, insertModel.ID, false, func(q *bun.SelectQuery) {
		q.Where("1 = 1")
	})
	assert.NoError(err)
	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "1 = 1")
}

func TestStore_FindByIDForUpdate_skip_locked(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	model, err := FindByIDForUpdate[testModel](context.Background(), db, insertModel.ID, true, nil)
	assert.NoError(err)
	assert.Equal(insertModel, model)

	assert.Len(qLogger.queries, 2)
	assert.Contains(qLogger.queries[1], "FOR UPDATE OF test_model SKIP LOCKED")
}

func TestStore_Save(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

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

func TestStore_Save_update(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	insertModel.Name = "foo"

	err = Save(context.Background(), db, insertModel, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestStore_Save_WithBeforeHooks_WithAfterHooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

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

	err := Save(context.Background(), db, insertModel, []BeforeHook[testModel]{beforeSave}, []AfterHook[testModel]{afterSave})
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
	assert.True(beforeSaveCalled)
	assert.True(afterSaveCalled)
}

func TestStore_Save_WithBeforeHooks_error(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	var beforeSaveErr = errors.New("test")
	beforeSave := func(ctx context.Context, db bun.IDB, model *testModel) error {
		return beforeSaveErr
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Save(context.Background(), db, insertModel, []BeforeHook[testModel]{beforeSave}, nil)
	assert.ErrorIs(err, beforeSaveErr)
}

func TestStore_Create(t *testing.T) {
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

	err := Create(context.Background(), db, insertModel, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Relation("Related").Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestStore_Create_exists(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Create(context.Background(), db, insertModel, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)

	err = Create(context.Background(), db, insertModel, nil, nil)
	pgErr := &pgdriver.Error{}
	assert.ErrorAs(err, pgErr)
	assert.True(pgErr.IntegrityViolation())
}

func TestStore_Create_WithBeforeHooks_WithAfterHooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	var beforeCreateCalled bool
	beforeCreate := func(ctx context.Context, db bun.IDB, model *testModel) error {
		beforeCreateCalled = true
		return nil
	}

	var afterCreateCalled bool
	afterCreate := func(ctx context.Context, model *testModel) {
		afterCreateCalled = true
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Create(context.Background(), db, insertModel, []BeforeHook[testModel]{beforeCreate}, []AfterHook[testModel]{afterCreate})
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
	assert.True(beforeCreateCalled)
	assert.True(afterCreateCalled)
}

func TestStore_Create_WithBeforeHooks_error(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	var beforeCreateErr = errors.New("test")
	beforeCreate := func(ctx context.Context, db bun.IDB, model *testModel) error {
		return beforeCreateErr
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Create(context.Background(), db, insertModel, []BeforeHook[testModel]{beforeCreate}, nil)
	assert.ErrorIs(err, beforeCreateErr)
}

func TestStore_Update(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	id := uuid.New().String()

	insertModel := &testModel{
		ID: id,
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	insertModel.Name = "foo"
	insertModel.Related = &testRelatedModel{
		ID:          uuid.New().String(),
		TestModelID: id,
	}

	err = Update(context.Background(), db, insertModel, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Relation("Related").Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
}

func TestStore_Update_not_exists(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	err := Update(context.Background(), db, insertModel, nil, nil)
	assert.ErrorIs(err, ErrUpdateNotExists)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestStore_Update_WithBeforeHooks_WithAfterHooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	var beforeUpdateCalled bool
	beforeUpdate := func(ctx context.Context, db bun.IDB, model *testModel) error {
		beforeUpdateCalled = true
		return nil
	}

	var afterUpdateCalled bool
	afterUpdate := func(ctx context.Context, model *testModel) {
		afterUpdateCalled = true
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Update(context.Background(), db, insertModel, []BeforeHook[testModel]{beforeUpdate}, []AfterHook[testModel]{afterUpdate})
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.NoError(err)

	assert.Equal(insertModel, model)
	assert.True(beforeUpdateCalled)
	assert.True(afterUpdateCalled)
}

func TestStore_Update_WithBeforeHooks_error(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	var beforeUpdateErr = errors.New("test")
	beforeUpdate := func(ctx context.Context, db bun.IDB, model *testModel) error {
		return beforeUpdateErr
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}

	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Update(context.Background(), db, insertModel, []BeforeHook[testModel]{beforeUpdate}, nil)
	assert.ErrorIs(err, beforeUpdateErr)
}

func TestStore_Delete(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel, nil, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestStore_Delete_query(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID:   uuid.New().String(),
		Name: "foo",
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel, func(q *bun.DeleteQuery) {
		q.Where("name = ?", "foo")
	}, nil, nil)
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestStore_Delete_WithBeforeHooks_WithAfterHooks(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

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

	err = Delete(context.Background(), db, insertModel, nil, []BeforeHook[testModel]{beforeDelete}, []AfterHook[testModel]{afterDelete})
	assert.NoError(err)

	model := &testModel{}
	err = db.NewSelect().Model(model).Scan(context.Background())
	assert.ErrorIs(err, sql.ErrNoRows)

	assert.True(beforeDeleteCalled)
	assert.True(afterDeleteCalled)
}

func TestStore_Delete_WithBeforeHooks_error(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	beforeDeleteErr := errors.New("test")
	beforeDelete := func(ctx context.Context, db bun.IDB, model *testModel) error {
		return beforeDeleteErr
	}

	insertModel := &testModel{
		ID: uuid.New().String(),
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	err = Delete(context.Background(), db, insertModel, nil, []BeforeHook[testModel]{beforeDelete}, nil)
	assert.ErrorIs(err, beforeDeleteErr)
}

func TestStore_Trx(t *testing.T) {
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

func TestStore_Trx_external_tx(t *testing.T) {
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

func TestStore_Trx_nested_transaction(t *testing.T) {
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
