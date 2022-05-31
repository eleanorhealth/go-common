package bao

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type BeforeHook[ModelT any] func(ctx context.Context, db bun.IDB, model *ModelT) error
type AfterHook[ModelT any] func(ctx context.Context, model *ModelT)

type Storer[ModelT any] interface {
	SelectQuery(ctx context.Context, model *ModelT) (*bun.SelectQuery, *schema.Table, error)
	SelectForUpdateQuery(ctx context.Context, model *ModelT, skipLocked bool) (*bun.SelectQuery, *schema.Table, error)
	Find(ctx context.Context, queryFn func(q *bun.SelectQuery)) ([]*ModelT, error)
	FindByID(ctx context.Context, id any) (*ModelT, error)
	FindByIDForUpdate(ctx context.Context, id any, skipLocked bool) (*ModelT, error)
	Save(ctx context.Context, model *ModelT) error
	Delete(ctx context.Context, model *ModelT) error

	Trx(ctx context.Context, fn func(ctx context.Context, txStore Storer[ModelT]) error) error

	WithBeforeSaveHooks(hooks ...BeforeHook[ModelT])
	WithAfterSaveHooks(hooks ...AfterHook[ModelT])
	WithBeforeDeleteHooks(hooks ...BeforeHook[ModelT])
	WithAfterDeleteHooks(hooks ...AfterHook[ModelT])
}

type Store[ModelT any] struct {
	db           bun.IDB
	beforeSave   []BeforeHook[ModelT]
	afterSave    []AfterHook[ModelT]
	beforeDelete []BeforeHook[ModelT]
	afterDelete  []AfterHook[ModelT]
}

var _ Storer[struct{}] = (*Store[struct{}])(nil)

func NewStore[ModelT any](db bun.IDB) *Store[ModelT] {
	return &Store[ModelT]{
		db: db,
	}
}

func (s *Store[ModelT]) SelectQuery(ctx context.Context, model *ModelT) (*bun.SelectQuery, *schema.Table, error) {
	return selectQuery(ctx, s.db, model)
}

func (s *Store[ModelT]) SelectForUpdateQuery(ctx context.Context, model *ModelT, skipLocked bool) (*bun.SelectQuery, *schema.Table, error) {
	return selectForUpdateQuery(ctx, s.db, model, skipLocked)
}

func (s *Store[ModelT]) Find(ctx context.Context, queryFn func(q *bun.SelectQuery)) ([]*ModelT, error) {
	return find[ModelT](ctx, s.db, queryFn)
}

func (s *Store[ModelT]) FindByID(ctx context.Context, id any) (*ModelT, error) {
	return findbyID[ModelT](ctx, s.db, id)
}

func (s *Store[ModelT]) FindByIDForUpdate(ctx context.Context, id any, skipLocked bool) (*ModelT, error) {
	return findbyIDForUpdate[ModelT](ctx, s.db, id, skipLocked)
}

func (s *Store[ModelT]) Save(ctx context.Context, model *ModelT) error {
	return save(ctx, s.db, model, s.beforeSave, s.afterSave)
}

func (s *Store[ModelT]) Delete(ctx context.Context, model *ModelT) error {
	return delete(ctx, s.db, model, s.beforeDelete, s.afterDelete)
}

func (s *Store[ModelT]) Trx(ctx context.Context, fn func(ctx context.Context, txStore Storer[ModelT]) error) error {
	return trx(ctx, s.db, func(ctx context.Context, tx bun.IDB) error {
		txStore := NewStore[ModelT](tx)
		return fn(ctx, txStore)
	})
}

func (s *Store[ModelT]) WithBeforeSaveHooks(hooks ...BeforeHook[ModelT]) {
	s.beforeSave = append(s.beforeSave, hooks...)
}

func (s *Store[ModelT]) WithAfterSaveHooks(hooks ...AfterHook[ModelT]) {
	s.afterSave = append(s.afterSave, hooks...)
}

func (s *Store[ModelT]) WithBeforeDeleteHooks(hooks ...BeforeHook[ModelT]) {
	s.beforeDelete = append(s.beforeDelete, hooks...)
}

func (s *Store[ModelT]) WithAfterDeleteHooks(hooks ...AfterHook[ModelT]) {
	s.afterDelete = append(s.afterDelete, hooks...)
}
