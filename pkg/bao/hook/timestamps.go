package hook

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/uptrace/bun"
)

func TimestampsBeforeHook[ModelT any](isCreate bool) Before[ModelT] {
	return func(ctx context.Context, db bun.IDB, model *ModelT) error {
		modelType := reflect.TypeOf(model).Elem()

		if modelType.Kind() != reflect.Struct {
			return errors.New("model must be a struct")
		}

		table := db.NewSelect().Model(model).DB().Table(modelType)

		now := time.Now().UTC()

		for _, field := range table.Fields {
			if field.StructField.Name != "CreatedAt" && field.StructField.Name != "UpdatedAt" {
				continue
			}

			if field.StructField.Name == "CreatedAt" && !isCreate {
				continue
			}

			modelValue := reflect.ValueOf(model).Elem()

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
