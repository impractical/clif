package clif

import (
	"context"
	"reflect"
)

// FlagValues holds the values of a flag specified at runtime. It's a list of
// [FlagValue]s, in the order they were specified.
//
// FlagValues gets its own type to enable normalization of multiple ways of
// specifying multiple values. For example, an application may want to accept
// both --flag=value1,value2,value3 and --flag=value1 --flag=value2
// --flag=value3 to set flag to ["value1", "value2", "value3"]. Making
// FlagValues a proper type allows clif to support normalizing those values.
type FlagValues []FlagValue

// As parses the values of the [FlagValues] into the target, which must be a
// pointer to a slice, a pointer to a type that implements [FlagValuesSetter],
// or a pointer that [FlagValue.As] knows how to parse into.
//
// If a pointer to a slice, the slice's element type must be a type that
// [FlagValue.As] can parse into. The slice will be populated with one value
// per [FlagValue], according to the rules of [FlagValue.As].
//
// If a pointer to a type that [FlagValue.As] can parse into *and* there's zero
// or one flag values passed during invocation, the parsing logic of
// [FlagValue.As] will be used. If no flag values are passed during invocation,
// the target will be set to its empty value. This is to simplify parsing a
// flag key that only accepts a single value, so the caller need not always
// deal with it as a slice. This approach is only recommended if the [FlagDef]
// prohibits multiple values for that flag key.
//
// If a pointer to a type that implements [FlagValuesSetter], the
// `SetFromFlagValues` method will be called and the [FlagValues] will be
// passed.
func (values FlagValues) As(ctx context.Context, target any) error {
	value := reflect.ValueOf(target)
	// the target has to be a pointer, so let's check that first
	if value.Kind() != reflect.Ptr {
		return InvalidKindError{
			Value: target,
			Kind:  value.Kind(),
		}
	}

	instantiated, err := newValues(ctx, values, value.Elem())
	if err != nil {
		return err
	}
	value.Elem().Set(instantiated)
	return nil
}
