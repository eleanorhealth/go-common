package bao2

import (
	"context"
	"database/sql"
	"errors"

	"github.com/eleanorhealth/go-common/pkg/bao2/hook"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/uptrace/bun"
)

type BaoStore[EntityT any, ModelT any] struct {
	db           bun.IDB
	fromEntityFn mapper[EntityT, ModelT]
	toEntityFn   mapper[ModelT, EntityT]
	relations    []string
	hooks        []hook.Hook
	notFoundErr  error
}

type mapper[FromT any, ToT any] func(from *FromT) (*ToT, error)

func NewBaoStore[EntityT any, ModelT any](db bun.IDB, fromEntity mapper[EntityT, ModelT], toEntity mapper[ModelT, EntityT], relations []string, hooks []hook.Hook, notFoundErr error) *BaoStore[EntityT, ModelT] {
	return &BaoStore[EntityT, ModelT]{
		db:           db,
		fromEntityFn: fromEntity,
		toEntityFn:   toEntity,
		relations:    relations,
		hooks:        hooks,
		notFoundErr:  notFoundErr,
	}
}

func (s *BaoStore[EntityT, ModelT]) toEntities(models []*ModelT) ([]*EntityT, error) {
	entities := make([]*EntityT, len(models))
	for i, model := range models {
		entity, err := s.toEntityFn(model)
		if err != nil {
			return nil, errs.Wrap(err, "model to entity")
		}

		entities[i] = entity
	}

	return entities, nil
}

func (s *BaoStore[EntityT, ModelT]) toEntity(model *ModelT, err error) (*EntityT, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, s.notFoundErr
		}

		return nil, errs.Wrap(err, "finding model")
	}

	entity, err := s.toEntityFn(model)
	if err != nil {
		return nil, errs.Wrap(err, "model to entity")
	}

	return entity, nil
}

func (s *BaoStore[EntityT, ModelT]) Find(ctx context.Context) ([]*EntityT, error) {
	var models []*ModelT

	q := s.db.NewSelect().Model(&models)

	for _, r := range s.relations {
		q.Relation(r)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return s.toEntities(models)
}

func (s *BaoStore[EntityT, ModelT]) FindByID(ctx context.Context, id any) (*EntityT, error) {
	var model ModelT

	q := s.db.NewSelect().Model(model)

	for _, r := range s.relations {
		q.Relation(r)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return s.toEntity(&model, err)
}

func (s *BaoStore[EntityT, ModelT]) FindByIDForUpdate(ctx context.Context, id any, skipLocked bool) (*EntityT, error) {
	var model ModelT

	q := s.db.NewSelect().Model(model)

	for _, r := range s.relations {
		q.Relation(r)
	}

	ForUpdate(ctx, q, skipLocked)

	err := q.Scan(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "scanning model")
	}

	return s.toEntity(&model, err)
}

func (s *BaoStore[EntityT, ModelT]) Create(ctx context.Context, entity *EntityT) error {
	model, err := s.fromEntityFn(entity)
	if err != nil {
		return errs.Wrap(err, "model from entity")
	}

	err = Create(ctx, s.db, model, s.hooks...)
	if err != nil {
		return errs.Wrap(err, "creating model")
	}

	return nil
}

func (s *BaoStore[EntityT, ModelT]) Update(ctx context.Context, entity *EntityT) error {
	model, err := s.fromEntityFn(entity)
	if err != nil {
		return errs.Wrap(err, "model from entity")
	}

	err = Create(ctx, s.db, model, s.hooks...)
	if err != nil {
		return errs.Wrap(err, "updating model")
	}

	return nil
}

func (s *BaoStore[EntityT, ModelT]) Delete(ctx context.Context, entity *EntityT) error {
	model, err := s.fromEntityFn(entity)
	if err != nil {
		return errs.Wrap(err, "model from entity")
	}

	err = Delete(ctx, s.db, model, s.hooks...)
	if err != nil {
		return errs.Wrap(err, "creating model")
	}

	return nil
}
