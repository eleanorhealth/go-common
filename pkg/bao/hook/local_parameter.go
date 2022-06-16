package hook

import (
	"context"
	"fmt"
	"strings"

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

		var builder strings.Builder
		var args []any

		for k, v := range vars {
			builder.WriteString(fmt.Sprintf(`SET LOCAL "%s" = ?;`, k))
			args = append(args, v)
		}

		_, err := db.ExecContext(ctx, builder.String(), args...)
		if err != nil {
			return errs.Wrapf(err, "setting local parameters")
		}

		return nil
	}
}
