package hook

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/uptrace/bun"
)

// TimestampsBeforeCreateHook sets CreatedAt and UpdatedAt to time.Now().UTC() before insert.
// Fields named CreatedAt or UpdatedAt must be of type time.Time to be populated; other types are skipped.
func TimestampsBeforeCreateHook[ModelT any]() Before[ModelT] {
	return timestampsBeforeHook[ModelT](true)
}

// TimestampsBeforeUpdateHook sets UpdatedAt to time.Now().UTC() before update.
// The field named UpdatedAt must be of type time.Time to be populated; other types are skipped.
func TimestampsBeforeUpdateHook[ModelT any]() Before[ModelT] {
	return timestampsBeforeHook[ModelT](false)
}

func timestampsBeforeHook[ModelT any](isCreate bool) Before[ModelT] {
	return func(ctx context.Context, db bun.IDB, model *ModelT) error {
		modelType := reflect.TypeOf(model).Elem()

		if modelType.Kind() != reflect.Struct {
			return errors.New("model must be a struct")
		}

		table := db.NewSelect().Model(model).DB().Table(modelType)

		now := time.Now().UTC()
		modelValue := reflect.ValueOf(model).Elem()

		for _, field := range table.Fields {
			if field.StructField.Name != "CreatedAt" && field.StructField.Name != "UpdatedAt" {
				continue
			}

			if field.StructField.Name == "CreatedAt" && !isCreate {
				continue
			}

			fieldValue := modelValue.FieldByIndex(field.StructField.Index)
			if !fieldValue.CanSet() {
				continue
			}

			switch fieldValue.Interface().(type) {
			case time.Time:
				fieldValue.Set(reflect.ValueOf(now))
			default:
				continue
			}
		}

		return nil
	}
}
