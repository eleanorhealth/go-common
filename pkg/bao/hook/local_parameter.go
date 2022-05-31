package hook

import (
	"context"
	"fmt"

	"github.com/eleanorhealth/go-common/pkg/bao"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/uptrace/bun"
)

func LocalParameterBeforeHook[ModelT any](fn func(ctx context.Context) map[string]string) bao.BeforeHook[ModelT] {
	return func(ctx context.Context, db bun.IDB, model *ModelT) error {
		var vars map[string]string

		if fn != nil {
			vars = fn(ctx)
		}

		for k, v := range vars {
			_, err := db.ExecContext(ctx, fmt.Sprintf(`SET LOCAL "%s" = ?`, k), v)
			if err != nil {
				return errs.Wrap(err, "setting local parameter")
			}
		}

		return nil
	}
}
