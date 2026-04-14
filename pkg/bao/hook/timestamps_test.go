package hook

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampsBeforeUpdateHook(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	hook := TimestampsBeforeUpdateHook[testModel]()

	createdAt := time.Now().Add(-time.Hour)
	model := &testModel{
		CreatedAt: createdAt,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	err := hook(context.Background(), db, model)
	assert.NoError(err)

	assert.WithinDuration(createdAt, model.CreatedAt, time.Second)
	assert.WithinDuration(time.Now(), model.UpdatedAt, time.Second)
	assert.Zero(model.OtherTime)
}

func TestTimestampsBeforeCreateHook(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	hook := TimestampsBeforeCreateHook[testModel]()

	model := &testModel{
		// CreatedAt is always overridden on create, even if pre-set.
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	err := hook(context.Background(), db, model)
	assert.NoError(err)

	assert.WithinDuration(time.Now(), model.CreatedAt, time.Second)
	assert.WithinDuration(time.Now(), model.UpdatedAt, time.Second)
	assert.Zero(model.OtherTime)
}
