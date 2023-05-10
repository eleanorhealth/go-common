package hook

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampsBeforeHook_update(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	hook := TimestampsBeforeHook[testModel](false)

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

func TestTimestampsBeforeHook_create(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	hook := TimestampsBeforeHook[testModel](true)

	model := &testModel{
		// Assuming these are preset for some reason
		// TimestampsBeforeHook will override them by design
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	err := hook(context.Background(), db, model)
	assert.NoError(err)

	assert.WithinDuration(time.Now(), model.CreatedAt, time.Second)
	assert.WithinDuration(time.Now(), model.UpdatedAt, time.Second)
	assert.Zero(model.OtherTime)
}
