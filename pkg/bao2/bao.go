package bao2

import (
	"context"
	"errors"
	"fmt"

	"github.com/eleanorhealth/go-common/pkg/bao2/hook"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/uptrace/bun"
)

func Create(ctx context.Context, db bun.IDB, model any, hooks ...hook.Hook) error {
	err := Trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		for _, h := range hooks {
			err := h.Before(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before hook")
			}
		}

		_, err := tx.NewInsert().Model(model).Exec(ctx)
		if err != nil {
			return errs.Wrap(err, "inserting model")
		}

		return nil
	})
	if err != nil {
		return errs.Wrap(err, "committing transaction")
	}

	for _, h := range hooks {
		h.After(ctx, model)
	}

	return nil
}

func Update(ctx context.Context, db bun.IDB, model any, hooks ...hook.Hook) error {
	err := Trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		for _, h := range hooks {
			err := h.Before(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before hook")
			}
		}

		_, err := tx.NewUpdate().Model(model).WherePK().Exec(ctx)
		if err != nil {
			return errs.Wrap(err, "updating model")
		}

		return nil
	})
	if err != nil {
		return errs.Wrap(err, "committing transaction")
	}

	for _, h := range hooks {
		h.After(ctx, model)
	}

	return nil
}

func Delete(ctx context.Context, db bun.IDB, model any, hooks ...hook.Hook) error {
	err := Trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		for _, h := range hooks {
			err := h.Before(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before hook")
			}
		}

		_, err := tx.NewDelete().Model(model).WherePK().Exec(ctx)
		if err != nil {
			return errs.Wrap(err, "deleting model")
		}

		return nil
	})
	if err != nil {
		return errs.Wrap(err, "committing transaction")
	}

	for _, h := range hooks {
		h.After(ctx, model)
	}

	return nil
}

func ForUpdate(ctx context.Context, query *bun.SelectQuery, skipLocked bool) *bun.SelectQuery {
	var skipLockedSQL string
	if skipLocked {
		skipLockedSQL = "SKIP LOCKED"
	}

	query.For(fmt.Sprintf("UPDATE OF %s %s", query.GetTableName(), skipLockedSQL))

	return query
}

func Trx(ctx context.Context, db bun.IDB, fn func(ctx context.Context, tx bun.IDB) error) error {
	var tx bun.Tx
	var err error
	var commit bool

	switch db := db.(type) {
	case bun.Tx:
		tx = db

	case *bun.DB:
		tx, err = db.BeginTx(ctx, nil)
		if err != nil {
			return errs.Wrap(err, "beginning transaction")
		}

		//nolint
		defer tx.Rollback()
		commit = true

	default:
		return errors.New("unknown type: %s")
	}

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	if commit {
		err = tx.Commit()
		if err != nil {
			return errs.Wrap(err, "committing transaction")
		}
	}

	return nil
}
