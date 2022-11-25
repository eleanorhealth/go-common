package hook

import (
	"context"
	"fmt"
	"strings"

	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/uptrace/bun"
)

type LocalParameter struct {
	vals func(context.Context) map[string]string
}

func NewLocalParameter(vals func(ctx context.Context) map[string]string) *LocalParameter {
	return &LocalParameter{
		vals: vals,
	}
}

func (l *LocalParameter) Before(ctx context.Context, db bun.IDB, model any) error {
	var vars map[string]string

	if l.vals != nil {
		vars = l.vals(ctx)
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

func (l *LocalParameter) After(ctx context.Context, model any) {
}
