package bao

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// testModelEntity represents a domain entity.
type testModelEntity struct {
	ID   string
	Name string
}

func testModelFromEntity(entity *testModelEntity) (*testModel, error) {
	return &testModel{
		ID:   entity.ID,
		Name: entity.Name,
	}, nil
}

func testModelToEntity(model *testModel) (*testModelEntity, error) {
	return &testModelEntity{
		ID:   model.ID,
		Name: model.Name,
	}, nil
}

func TestNewStore(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	notFoundErr := errors.New("not found")
	store := NewStore(db, testModelFromEntity, testModelToEntity, notFoundErr, nil, nil, nil, nil, nil)

	assert.Equal(db, store.db)
	assert.NotNil(store.fromEntityFn)
	assert.NotNil(store.toEntityFn)
	assert.Equal(notFoundErr, store.notFoundError)
}

func TestStore_Find(t *testing.T) {
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

	notFoundErr := errors.New("not found")
	store := NewStore(db, testModelFromEntity, testModelToEntity, notFoundErr, nil, nil, nil, nil, nil)

	model, err := store.Find(context.Background())
	assert.NoError(err)

	assert.Len(model, 2)
}

func TestStore_FindByID(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID:   uuid.New().String(),
		Name: "foo",
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	notFoundErr := errors.New("not found")
	store := NewStore(db, testModelFromEntity, testModelToEntity, notFoundErr, nil, nil, nil, nil, nil)

	model, err := store.FindByID(context.Background(), insertModel.ID)
	assert.NoError(err)
	assert.Equal(insertModel.ID, model.ID)

	model, err = store.FindByID(context.Background(), uuid.New().String())
	assert.ErrorIs(err, notFoundErr)
	assert.Nil(model)
}

func TestStore_FindByIDForUpdate(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	insertModel := &testModel{
		ID:   uuid.New().String(),
		Name: "foo",
	}
	_, err := db.NewInsert().Model(insertModel).Exec(context.Background())
	assert.NoError(err)

	notFoundErr := errors.New("not found")
	store := NewStore(db, testModelFromEntity, testModelToEntity, notFoundErr, nil, nil, nil, nil, nil)

	model, err := store.FindByIDForUpdate(context.Background(), insertModel.ID, false)
	assert.NoError(err)
	assert.Equal(insertModel.ID, model.ID)

	model, err = store.FindByIDForUpdate(context.Background(), insertModel.ID, true)
	assert.NoError(err)
	assert.Equal(insertModel.ID, model.ID)

	model, err = store.FindByIDForUpdate(context.Background(), uuid.New().String(), false)
	assert.ErrorIs(err, notFoundErr)
	assert.Nil(model)
}

func TestStore_Create(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	entity := &testModelEntity{
		ID:   uuid.New().String(),
		Name: "foo",
	}

	notFoundErr := errors.New("not found")
	store := NewStore(db, testModelFromEntity, testModelToEntity, notFoundErr, nil, nil, nil, nil, nil)

	err := store.Create(context.Background(), entity)
	assert.NoError(err)

	foundEntity, err := store.FindByID(context.Background(), entity.ID)
	assert.NoError(err)
	assert.Equal(entity, foundEntity)
}

func TestStore_Update(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	entity := &testModelEntity{
		ID:   uuid.New().String(),
		Name: "foo",
	}

	notFoundErr := errors.New("not found")
	store := NewStore(db, testModelFromEntity, testModelToEntity, notFoundErr, nil, nil, nil, nil, nil)

	err := store.Create(context.Background(), entity)
	assert.NoError(err)

	entity.Name = "baz"

	err = store.Update(context.Background(), entity)
	assert.NoError(err)

	foundEntity, err := store.FindByID(context.Background(), entity.ID)
	assert.NoError(err)
	assert.Equal(entity, foundEntity)

	err = store.Update(context.Background(), &testModelEntity{
		ID: uuid.New().String(),
	})
	assert.ErrorIs(err, notFoundErr)
}

func TestStore_Delete(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	entity := &testModelEntity{
		ID:   uuid.New().String(),
		Name: "foo",
	}

	notFoundErr := errors.New("not found")
	store := NewStore(db, testModelFromEntity, testModelToEntity, notFoundErr, nil, nil, nil, nil, nil)

	err := store.Create(context.Background(), entity)
	assert.NoError(err)

	foundEntity, err := store.FindByID(context.Background(), entity.ID)
	assert.NoError(err)
	assert.Equal(entity, foundEntity)

	err = store.Delete(context.Background(), entity)
	assert.NoError(err)

	foundEntity, err = store.FindByID(context.Background(), entity.ID)
	assert.ErrorIs(err, notFoundErr)
	assert.Nil(foundEntity)
}
