package bao

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type BeforeHook[ModelT any] func(ctx context.Context, db bun.IDB, model *ModelT) error
type AfterHook[ModelT any] func(ctx context.Context, model *ModelT)

func SelectQuery[ModelT any](ctx context.Context, db bun.IDB, model *ModelT) (*bun.SelectQuery, *schema.Table, error) {
	rType := reflect.TypeOf(model)
	if rType.Elem().Kind() != reflect.Struct && rType.Elem().Kind() != reflect.Slice {
		return nil, nil, ErrModelNotStructSlicePointer
	}

	query := db.NewSelect().Model(model)
	var table *schema.Table

	if rType.Elem().Kind() == reflect.Struct {
		table = query.DB().Table(rType)
	} else {
		table = query.DB().Table(rType.Elem().Elem())
	}

	return query, table, nil
}

func SelectForUpdateQuery[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, skipLocked bool) (*bun.SelectQuery, *schema.Table, error) {
	query, table, err := SelectQuery(ctx, db, model)
	if err != nil {
		return nil, nil, err
	}

	var skipLockedSQL string
	if skipLocked {
		skipLockedSQL = "SKIP LOCKED"
	}

	query.For(fmt.Sprintf("UPDATE OF %s %s", table.Alias, skipLockedSQL))

	return query, table, nil
}

func Find[ModelT any](ctx context.Context, db bun.IDB, queryFn func(q *bun.SelectQuery)) ([]*ModelT, error) {
	var model []*ModelT
	query, _, err := SelectQuery(ctx, db, &model)
	if err != nil {
		return nil, errs.Wrap(err, "select query")
	}

	if queryFn != nil {
		queryFn(query)
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return model, nil
}

func FindFirst[ModelT any](ctx context.Context, db bun.IDB, queryFn func(q *bun.SelectQuery)) (*ModelT, error) {
	var model ModelT
	query, _, err := SelectQuery(ctx, db, &model)
	if err != nil {
		return nil, errs.Wrap(err, "select query")
	}

	if queryFn != nil {
		queryFn(query)
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return &model, nil
}

func FindByID[ModelT any](ctx context.Context, db bun.IDB, id any) (*ModelT, error) {
	var model ModelT
	query, table, err := SelectQuery(ctx, db, &model)
	if err != nil {
		return nil, errs.Wrap(err, "select query")
	}

	if len(table.PKs) != 1 {
		return nil, ErrOnePrimaryKey
	}

	query.Where(fmt.Sprintf("%s.%s = ?", table.SQLAlias, table.PKs[0].SQLName), id)

	err = query.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return &model, nil
}

func FindByIDForUpdate[ModelT any](ctx context.Context, db bun.IDB, id any, skipLocked bool) (*ModelT, error) {
	var model ModelT
	query, table, err := SelectForUpdateQuery(ctx, db, &model, skipLocked)
	if err != nil {
		return nil, errs.Wrap(err, "select for update query")
	}

	if len(table.PKs) != 1 {
		return nil, ErrOnePrimaryKey
	}

	query.Where(fmt.Sprintf("%s.%s = ?", table.SQLAlias, table.PKs[0].SQLName), id)

	err = query.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return &model, nil
}

func Save[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, befores []BeforeHook[ModelT], afters []AfterHook[ModelT]) error {
	rType := reflect.TypeOf(model)
	if rType.Kind() != reflect.Ptr {
		return ErrModelNotPointer
	}

	if rType.Elem().Kind() != reflect.Struct && rType.Elem().Kind() != reflect.Slice {
		return ErrModelNotStructSlicePointer
	}

	err := Trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		for _, fn := range befores {
			err := fn(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before save hook")
			}
		}

		exists, err := tx.NewSelect().Model(model).WherePK().Exists(ctx)
		if err != nil {
			return errs.Wrap(err, "checking if model exists")
		}

		if !exists {
			_, err := tx.NewInsert().Model(model).Exec(ctx)
			if err != nil {
				return errs.Wrap(err, "inserting model")
			}
		} else {
			_, err := tx.NewUpdate().Model(model).WherePK().Exec(ctx)
			if err != nil {
				return errs.Wrap(err, "updating model")
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	for _, fn := range afters {
		fn(ctx, model)
	}

	return nil
}

func Delete[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, queryFn func(q *bun.DeleteQuery), befores []BeforeHook[ModelT], afters []AfterHook[ModelT]) error {
	rType := reflect.TypeOf(model)
	if rType.Kind() != reflect.Ptr {
		return ErrModelNotPointer
	}

	if rType.Elem().Kind() != reflect.Struct && rType.Elem().Kind() != reflect.Slice {
		return ErrModelNotStructSlicePointer
	}

	err := Trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		for _, fn := range befores {
			err := fn(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before delete hook")
			}
		}

		query := tx.NewDelete().Model(model)

		if queryFn != nil {
			queryFn(query)
		} else {
			query.WherePK()
		}

		_, err := query.Exec(ctx)
		if err != nil {
			return errs.Wrap(err, "deleting model")
		}

		return nil
	})
	if err != nil {
		return err
	}

	for _, fn := range afters {
		fn(ctx, model)
	}

	return nil
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
