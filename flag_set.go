package clif

import (
	"context"
	"reflect"
)

// FlagSet holds the [FlagValues] that were specified for each flag key in the
// invocation. It's its own type so we can map an entire specification of flags
// onto a struct using reflection, if we want.
type FlagSet map[string]FlagValues

// As parses the values of the [FlagSet] into the target, which must be a
// pointer to a struct, a pointer to a map, or a pointer to a type that
// implements [FlagSetSetter].
//
// If a pointer to a map, the flag keys will be used as the keys of the map.
// All flag values must be parsed into the same type, and will be parsed into
// the types supported by [FlagValues.As].
//
// If a pointer to a struct, the `flag` struct tag will control which fields a
// flag key parses into. The value of the tag should be the canonical name
// (i.e., not an alias) of the flag that should be parsed into that field.
// Unexported struct fields cannot be parsed into. All exported struct fields
// must have a `flag` struct tag; use "-" as a struct tag to not parse into a
// field. Field types can be any type supported by [FlagValues.As].
//
// If a pointer to a type that implements [FlagSetSetter], the `SetFromFlagSet`
// method will be called and the [FlagSet] will be passed.
func (set FlagSet) As(ctx context.Context, target any) error {
	value := reflect.ValueOf(target)
	// the target has to be a pointer, so let's check that first
	if value.Kind() != reflect.Ptr {
		return InvalidKindError{
			Value: target,
			Kind:  value.Kind(),
		}
	}

	instantiated, err := newSet(ctx, set, value.Elem())
	if err != nil {
		return err
	}
	value.Elem().Set(instantiated)
	return nil
}
