package bao

import (
	"context"

	"github.com/uptrace/bun"
)

type BeforeHook[ModelT any] func(ctx context.Context, db bun.IDB, model *ModelT) error
type AfterHook[ModelT any] func(ctx context.Context, model *ModelT)
