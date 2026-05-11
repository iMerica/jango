package orm

import (
	"context"
	"reflect"
)

type QuerySet[T any] struct {
	*BaseQuerySet
}

func NewQuerySet[T any](model *ModelMeta, db *DB) *QuerySet[T] {
	return &QuerySet[T]{
		BaseQuerySet: NewBaseQuerySet(model, db),
	}
}

func Objects[T any](appLabel, modelName string) *QuerySet[T] {
	meta := GlobalRegistry().MustGet(appLabel, modelName)
	return NewQuerySet[T](meta, defaultDB)
}

func (qs *QuerySet[T]) Filter(lookups ...Lookup) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.Filter(lookups...)}
}

func (qs *QuerySet[T]) FilterQ(q QNode) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.FilterQ(q)}
}

func (qs *QuerySet[T]) Exclude(lookups ...Lookup) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.Exclude(lookups...)}
}

func (qs *QuerySet[T]) ExcludeQ(q QNode) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.ExcludeQ(q)}
}

func (qs *QuerySet[T]) OrderBy(fields ...string) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.OrderBy(fields...)}
}

func (qs *QuerySet[T]) Limit(n int) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.Limit(n)}
}

func (qs *QuerySet[T]) Offset(n int) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.Offset(n)}
}

func (qs *QuerySet[T]) SelectRelated(fields ...string) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.SelectRelated(fields...)}
}

func (qs *QuerySet[T]) PrefetchRelated(field string, opts ...PrefetchOption) *QuerySet[T] {
	return &QuerySet[T]{BaseQuerySet: qs.BaseQuerySet.PrefetchRelated(field, opts...)}
}

func (qs *QuerySet[T]) First(ctx context.Context) (*T, error) {
	limited := qs.BaseQuerySet.OrderBy(qs.model.PKField).Limit(1)
	rows, err := limited.executeSelectRows(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := scanRowsGeneric[T](rows, qs.model)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (qs *QuerySet[T]) Get(ctx context.Context, lookups ...Lookup) (*T, error) {
	filtered := qs.BaseQuerySet.Filter(lookups...)
	rows, err := filtered.executeSelectRows(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := scanRowsGeneric[T](rows, qs.model)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, &DoesNotExist{ModelName: qs.model.ModelName, Filter: formatLookups(lookups)}
	}
	if len(results) > 1 {
		return nil, &MultipleObjectsReturned{ModelName: qs.model.ModelName, Filter: formatLookups(lookups)}
	}
	return results[0], nil
}

func (qs *QuerySet[T]) AllRecords(ctx context.Context) ([]*T, error) {
	rows, err := qs.BaseQuerySet.executeSelectRows(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsGeneric[T](rows, qs.model)
}

func (qs *QuerySet[T]) Create(ctx context.Context, instance *T) error {
	values := extractValuesFromStruct(instance, qs.model)
	createdMap, err := qs.BaseQuerySet.Create(ctx, values)
	if err != nil {
		return err
	}
	// We'd ideally scan the map back into the instance or let the caller re-fetch.
	// For simplicity, we can set the PK.
	setPKFromMap(instance, qs.model, createdMap)
	return nil
}

func scanRowsGeneric[T any](rows Rows, meta *ModelMeta) ([]*T, error) {
	var results []*T
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var instance T
		val := reflect.ValueOf(&instance).Elem()

		// Prepare pointers to pass to Scan
		scanArgs := make([]interface{}, len(cols))
		for i, col := range cols {
			// Find field by db column
			var fieldName string
			for _, f := range meta.Fields {
				if meta.DBColumnForField(f.Name) == col {
					fieldName = f.Name
					break
				}
			}

			if fieldName != "" {
				structField := val.FieldByName(fieldName)
				if structField.IsValid() && structField.CanAddr() {
					scanArgs[i] = structField.Addr().Interface()
					continue
				}
			}
			// Fallback: dummy scanner
			var dummy interface{}
			scanArgs[i] = &dummy
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		results = append(results, &instance)
	}

	return results, rows.Err()
}

func extractValuesFromStruct(instance interface{}, meta *ModelMeta) map[string]interface{} {
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	values := make(map[string]interface{})
	for _, f := range meta.Fields {
		structField := val.FieldByName(f.Name)
		if structField.IsValid() {
			if !f.PrimaryKey || !f.Auto || !structField.IsZero() {
				values[f.Name] = structField.Interface()
			}
		}
	}
	return values
}

func setPKFromMap(instance interface{}, meta *ModelMeta, data map[string]interface{}) {
	if meta.PKField == "" {
		return
	}
	pkVal, ok := data[meta.PKField]
	if !ok {
		pkVal, ok = data[meta.PKColumn()]
		if !ok {
			return
		}
	}
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	structField := val.FieldByName(meta.PKField)
	if structField.IsValid() && structField.CanSet() {
		// handle basic type conversions
		rv := reflect.ValueOf(pkVal)
		if rv.Type().ConvertibleTo(structField.Type()) {
			structField.Set(rv.Convert(structField.Type()))
		}
	}
}
