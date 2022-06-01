package hook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalParameterBeforeHook(t *testing.T) {
	assert := assert.New(t)

	db := testDB(t)

	qLogger := &queryLogger{}
	db.AddQueryHook(qLogger)

	key := "myapp.foo"
	value := "bar"

	hook := LocalParameterBeforeHook[testModel](func(ctx context.Context) map[string]string {
		return map[string]string{
			key: value,
		}
	})

	err := hook(context.Background(), db, &testModel{})
	assert.NoError(err)

	assert.Len(qLogger.queries, 1)
	assert.Contains(qLogger.queries[0], `SET LOCAL "myapp.foo" = 'bar'`)
}
