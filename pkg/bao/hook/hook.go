package hook

import (
	"context"

	"github.com/uptrace/bun"
)

type Before[ModelT any] func(ctx context.Context, db bun.IDB, model *ModelT) error
type After[ModelT any] func(ctx context.Context, model *ModelT)
