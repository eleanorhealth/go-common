package bao

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/eleanorhealth/go-common/pkg/bao/hook"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/fatih/structtag"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

var ErrUpdateNotExists = errors.New("model to be updated does not exist")

type relatedModelOp int

const (
	relatedModelOpUpdate relatedModelOp = iota
	relatedModelOpDelete relatedModelOp = iota
)

func SelectQuery[ModelT any](ctx context.Context, db bun.IDB, model *ModelT) (*bun.SelectQuery, *schema.Table, error) {
	rType := reflect.TypeOf(model)
	if rType.Elem().Kind() != reflect.Struct && rType.Elem().Kind() != reflect.Slice {
		return nil, nil, ErrModelNotStructOrSlice
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

func FindByID[ModelT any](ctx context.Context, db bun.IDB, id any, queryFn func(q *bun.SelectQuery)) (*ModelT, error) {
	var model ModelT
	query, table, err := SelectQuery(ctx, db, &model)
	if err != nil {
		return nil, errs.Wrap(err, "select query")
	}

	if len(table.PKs) != 1 {
		return nil, ErrOnePrimaryKey
	}

	query.Where(fmt.Sprintf("%s.%s = ?", table.SQLAlias, table.PKs[0].SQLName), id)

	if queryFn != nil {
		queryFn(query)
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return &model, nil
}

func FindByIDForUpdate[ModelT any](ctx context.Context, db bun.IDB, id any, skipLocked bool, queryFn func(q *bun.SelectQuery)) (*ModelT, error) {
	var model ModelT
	query, table, err := SelectForUpdateQuery(ctx, db, &model, skipLocked)
	if err != nil {
		return nil, errs.Wrap(err, "select for update query")
	}

	if len(table.PKs) != 1 {
		return nil, ErrOnePrimaryKey
	}

	query.Where(fmt.Sprintf("%s.%s = ?", table.SQLAlias, table.PKs[0].SQLName), id)

	if queryFn != nil {
		queryFn(query)
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return &model, nil
}

func Create[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, befores []hook.Before[ModelT], afters []hook.After[ModelT]) error {
	rType := reflect.TypeOf(model)
	if rType.Elem().Kind() != reflect.Struct {
		return ErrModelNotStruct
	}

	err := Trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		for _, fn := range befores {
			err := fn(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before save hook")
			}
		}

		_, err := tx.NewInsert().Model(model).Exec(ctx)
		if err != nil {
			return errs.Wrap(err, "inserting model")
		}

		err = relatedModels(ctx, tx, model, relatedModelOpUpdate)
		if err != nil {
			return errs.Wrap(err, "creating related models")
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

func Update[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, befores []hook.Before[ModelT], afters []hook.After[ModelT]) error {
	rType := reflect.TypeOf(model)
	if rType.Elem().Kind() != reflect.Struct {
		return ErrModelNotStruct
	}

	err := Trx(ctx, db, func(ctx context.Context, tx bun.IDB) error {
		exists, err := tx.NewSelect().Model(model).WherePK().Exists(ctx)
		if err != nil {
			return errs.Wrap(err, "checking if model exists")
		}

		if !exists {
			return ErrUpdateNotExists
		}

		for _, fn := range befores {
			err := fn(ctx, tx, model)
			if err != nil {
				return errs.Wrap(err, "before save hook")
			}
		}

		_, err = tx.NewUpdate().Model(model).WherePK().Exec(ctx)
		if err != nil {
			return errs.Wrap(err, "updating model")
		}

		err = relatedModels(ctx, tx, model, relatedModelOpUpdate)
		if err != nil {
			return errs.Wrap(err, "updating related models")
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

func Delete[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, queryFn func(q *bun.DeleteQuery), befores []hook.Before[ModelT], afters []hook.After[ModelT]) error {
	rType := reflect.TypeOf(model)
	if rType.Elem().Kind() != reflect.Struct {
		return ErrModelNotStruct
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

		err = relatedModels(ctx, tx, model, relatedModelOpDelete)
		if err != nil {
			return errs.Wrap(err, "deleting related models")
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

func relatedModels[ModelT any](ctx context.Context, bun bun.IDB, model *ModelT, op relatedModelOp) error {
	modelType := reflect.TypeOf(model).Elem()

	if modelType.Kind() != reflect.Struct {
		return ErrModelNotStruct
	}

	table := bun.NewSelect().Model(model).DB().Table(modelType)

	for _, relation := range table.Relations {
		idx := relation.Field.StructField.Index
		tag := modelType.FieldByIndex(idx).Tag

		tags, err := structtag.Parse(string(tag))
		if err != nil {
			return errs.Wrap(err, "parsing tags")
		}

		baoTag, err := tags.Get("bao")
		if err != nil {
			continue
		}

		if !baoTag.HasOption("persist") {
			continue
		}

		deleteModel := reflect.New(relation.JoinTable.Type)
		q := bun.NewDelete().Model(deleteModel.Interface())

		for _, joinField := range relation.JoinFields {
			baseField := relation.BaseFields[0]
			q.Where(fmt.Sprintf("%s = ?", string(joinField.SQLName)), baseField.Value(reflect.ValueOf(*model)).Interface())
		}

		_, err = q.Exec(ctx)
		if err != nil {
			return errs.Wrapf(err, "deleting related model (%s)", relation.JoinTable.ModelName)
		}

		// Return early if the operation is to delete.
		if op == relatedModelOpDelete {
			return nil
		}

		insertModel := relation.Field.Value(reflect.ValueOf(*model))
		if insertModel.IsZero() {
			continue
		}

		_, err = bun.NewInsert().Model(insertModel.Interface()).Exec(ctx)
		if err != nil {
			return errs.Wrapf(err, "inserting related model (%s)", relation.JoinTable.ModelName)
		}
	}

	return nil
}
