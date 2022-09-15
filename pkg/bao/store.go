package bao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/eleanorhealth/go-common/pkg/bao/hook"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"github.com/uptrace/bun"
)

type Store[EntityT any, ModelT any] struct {
	db            bun.IDB
	fromEntityFn  mapper[EntityT, ModelT]
	toEntityFn    mapper[ModelT, EntityT]
	notFoundError error
	relations     []string

	beforeCreateUpdateHooks []hook.Before[ModelT]
	afterCreateUpdateHooks  []hook.After[ModelT]
	beforeDeleteHooks       []hook.Before[ModelT]
	afterDeleteHooks        []hook.After[ModelT]
}

type mapper[FromT any, ToT any] func(from *FromT) (*ToT, error)

func NewStore[EntityT any, ModelT any](db bun.IDB, fromEntity mapper[EntityT, ModelT], toEntity mapper[ModelT, EntityT], notFoundError error, relations []string, beforeCreateUpdateHooks []hook.Before[ModelT], afterCreateUpdateHooks []hook.After[ModelT], beforeDeleteHooks []hook.Before[ModelT], afterDeleteHooks []hook.After[ModelT]) *Store[EntityT, ModelT] {
	return &Store[EntityT, ModelT]{
		db:                      db,
		fromEntityFn:            fromEntity,
		toEntityFn:              toEntity,
		notFoundError:           notFoundError,
		relations:               relations,
		beforeCreateUpdateHooks: beforeCreateUpdateHooks,
		afterCreateUpdateHooks:  afterCreateUpdateHooks,
		beforeDeleteHooks:       beforeDeleteHooks,
		afterDeleteHooks:        afterDeleteHooks,
	}
}

func (s *Store[EntityT, ModelT]) toEntities(models []*ModelT) ([]*EntityT, error) {
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

func (s *Store[EntityT, ModelT]) toEntity(model *ModelT, err error) (*EntityT, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, s.notFoundError
		}

		return nil, errs.Wrap(err, "finding model")
	}

	entity, err := s.toEntityFn(model)
	if err != nil {
		return nil, errs.Wrap(err, "model to entity")
	}

	return entity, nil
}

func (s *Store[EntityT, ModelT]) Find(ctx context.Context) ([]*EntityT, error) {
	models, err := Find[ModelT](ctx, s.db, func(q *bun.SelectQuery) {
		for _, relation := range s.relations {
			q.Relation(relation)
		}
	})
	if err != nil {
		return nil, errs.Wrap(err, "finding model")
	}

	return s.toEntities(models)
}

func (s *Store[EntityT, ModelT]) FindByID(ctx context.Context, id any) (*EntityT, error) {
	model, err := FindByID[ModelT](ctx, s.db, id, func(q *bun.SelectQuery) {
		for _, relation := range s.relations {
			q.Relation(relation)
		}
	})

	return s.toEntity(model, err)
}

func (s *Store[EntityT, ModelT]) FindByIDForUpdate(ctx context.Context, id any, skipLocked bool) (*EntityT, error) {
	model, err := FindByIDForUpdate[ModelT](ctx, s.db, id, skipLocked, func(q *bun.SelectQuery) {
		for _, relation := range s.relations {
			q.Relation(relation)
		}
	})

	return s.toEntity(model, err)
}

func (s *Store[EntityT, ModelT]) Create(ctx context.Context, entity *EntityT) error {
	model, err := s.fromEntityFn(entity)
	if err != nil {
		return errs.Wrap(err, "model from entity")
	}

	err = Create(
		ctx,
		s.db,
		model,
		s.beforeCreateUpdateHooks,
		s.afterCreateUpdateHooks,
	)
	if err != nil {
		return errs.Wrap(err, "creating model")
	}

	return nil
}

func (s *Store[EntityT, ModelT]) Update(ctx context.Context, entity *EntityT) error {
	model, err := s.fromEntityFn(entity)
	if err != nil {
		return errs.Wrap(err, "model from entity")
	}

	err = Update(
		ctx,
		s.db,
		model,
		s.beforeCreateUpdateHooks,
		s.afterCreateUpdateHooks,
	)
	if err != nil {
		if errors.Is(err, ErrUpdateNotExists) {
			return s.notFoundError
		}

		return errs.Wrap(err, "updating model")
	}

	return nil
}

func (s *Store[EntityT, ModelT]) Delete(ctx context.Context, entity *EntityT) error {
	model, err := s.fromEntityFn(entity)
	if err != nil {
		return errs.Wrap(err, "model from entity")
	}

	err = Delete(
		ctx,
		s.db,
		model,
		nil,
		s.beforeDeleteHooks,
		s.afterDeleteHooks,
	)
	if err != nil {
		return errs.Wrap(err, "deleting model")
	}

	return nil
}
