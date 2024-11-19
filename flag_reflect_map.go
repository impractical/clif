package clif

import (
	"context"
	"reflect"
)

func newSetFromMap(ctx context.Context, flags FlagSet, target reflect.Value) (reflect.Value, error) {
	elemType := target.Type().Elem()
	results := reflect.MakeMap(target.Type())

	for key, values := range flags {
		elemValue := reflect.Zero(elemType)
		element, err := newValues(ctx, values, elemValue)
		if err != nil {
			return target, err
		}
		results.SetMapIndex(reflect.ValueOf(key), element)
	}
	return results, nil
}
