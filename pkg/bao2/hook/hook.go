package hook

import (
	"context"

	"github.com/uptrace/bun"
)

type Hook interface {
	Before(ctx context.Context, db bun.IDB, model any) error
	After(ctx context.Context, model any)
}
