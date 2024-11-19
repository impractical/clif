package clif

import (
	"context"
	"reflect"
)

func newValues(ctx context.Context, values FlagValues, target reflect.Value) (reflect.Value, error) {
	if target.Type().Implements(reflect.TypeOf((*FlagValuesSetter)(nil)).Elem()) {
		return newValuesFromValuesSetter(ctx, values, target)
	}

	// if we're pointing to a pointer, dereference it and try again on
	// whatever the pointer is pointing to
	if target.Kind() == reflect.Ptr {
		return newValuesFromPointer(ctx, values, target)
	}

	// if the pointer isn't pointing to a slice and we have multiple
	// values, that's not going to work
	if target.Kind() != reflect.Slice && len(values) > 1 {
		return target, InvalidKindError{
			Value: target,
			Kind:  target.Kind(),
		}
	}

	// if the pointer isn't pointing to a slice and we have only one value,
	// just convert the value into the target
	if target.Kind() != reflect.Slice && len(values) == 1 {
		return newValue(ctx, values[0], target)
	}

	// if we don't have any values, we want the zero value of whatever
	// we're pointing at, be it a slice or something else
	if len(values) < 1 && target.Kind() != reflect.Slice {
		return newValue(ctx, FlagValue{}, target)
	}

	elemType := target.Type().Elem()
	results := reflect.MakeSlice(target.Type(), 0, len(values))

	for _, flag := range values {
		elemTarget := reflect.Zero(elemType)
		element, err := newValue(ctx, flag, elemTarget)
		if err != nil {
			return target, err
		}
		results = reflect.Append(results, element)
	}

	return results, nil
}
