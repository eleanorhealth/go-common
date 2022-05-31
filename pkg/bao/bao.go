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

func selectQuery(ctx context.Context, db bun.IDB, model any) (*bun.SelectQuery, *schema.Table, error) {
	rType := reflect.TypeOf(model)
	if rType.Kind() != reflect.Ptr {
		return nil, nil, ErrModelNotPointer
	}

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

	relations(query, table, "")

	return query, table, nil
}

func selectForUpdateQuery(ctx context.Context, db bun.IDB, model any, skipLocked bool) (*bun.SelectQuery, *schema.Table, error) {
	query, table, err := selectQuery(ctx, db, model)
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

func find[ModelT any](ctx context.Context, db bun.IDB, queryFn func(q *bun.SelectQuery)) ([]*ModelT, error) {
	var model []*ModelT
	query, _, err := selectQuery(ctx, db, &model)
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

func findbyID[ModelT any](ctx context.Context, db bun.IDB, id any) (*ModelT, error) {
	var model ModelT
	query, table, err := selectQuery(ctx, db, &model)
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

func findbyIDForUpdate[ModelT any](ctx context.Context, db bun.IDB, id any, skipLocked bool) (*ModelT, error) {
	var model ModelT
	query, table, err := selectForUpdateQuery(ctx, db, &model, skipLocked)
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

func save[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, befores []BeforeHook[ModelT], afters []AfterHook[ModelT]) error {
	rType := reflect.TypeOf(model)
	if rType.Kind() != reflect.Ptr {
		return ErrModelNotPointer
	}

	if rType.Elem().Kind() != reflect.Struct && rType.Elem().Kind() != reflect.Slice {
		return ErrModelNotStructSlicePointer
	}

	err := trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
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

		for _, fn := range afters {
			fn(ctx, model)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func delete[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, befores []BeforeHook[ModelT], afters []AfterHook[ModelT]) error {
	rType := reflect.TypeOf(model)
	if rType.Kind() != reflect.Ptr {
		return ErrModelNotPointer
	}

	if rType.Elem().Kind() != reflect.Struct && rType.Elem().Kind() != reflect.Slice {
		return ErrModelNotStructSlicePointer
	}

	err := trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		for _, fn := range befores {
			err := fn(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before delete hook")
			}
		}

		_, err := tx.NewDelete().Model(model).WherePK().Exec(ctx)
		if err != nil {
			return errs.Wrap(err, "deleting model")
		}

		for _, fn := range afters {
			fn(ctx, model)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func relations(query *bun.SelectQuery, table *schema.Table, parent string) {
	for _, relation := range table.Relations {
		var relationName string
		if len(parent) > 0 {
			relationName = parent + "." + relation.Field.GoName
		} else {
			relationName = relation.Field.GoName
		}

		query.Relation(relationName)

		if relation.JoinTable != nil {
			relations(query, relation.JoinTable, relationName)
		}
	}
}

func trx(ctx context.Context, db bun.IDB, fn func(ctx context.Context, tx bun.IDB) error) error {
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
