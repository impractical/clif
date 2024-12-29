package clif

import (
	"context"
	"reflect"
)

// FlagValue holds the value of a flag specified at runtime.
type FlagValue struct {
	// HasValue indicates whether the flag had a value set. If false, it
	// indicates that the flag was used as a toggle, without a value, i.e.
	// --flag. If true, it indicates the flag was used with a value, i.e.
	// --flag=value or --flag value.
	HasValue bool

	// Raw holds the value the flag was given, if Set is true.
	Raw string

	// Key is the key used to invoke the flag, which could be an alias.
	Key string

	// CanonicalKey is the Name of the flag in the FlagDef, the canonical
	// way to refer to the flag.
	CanonicalKey string
}

// As parses the value of the [FlagValue] into the target, which must be a
// pointer to a boolean, a pointer to a string, a pointer to an integer, a
// pointer to an unsigned integer, a pointer to a float, a pointer to a
// [time.Time], a pointer to a [*big.Int], a pointer to a [*big.Float], or a
// pointer to a type that implements [FlagValueSetter].
//
// For a boolean, [strconv.ParseBool] will be used to parse the value, unless
// [FlagValue.Set] is false, in which case the value will be `true`.
//
// For a string, no conversion will be done and [FlagValue.Raw] will be
// returned.
//
// For an integer, [strconv.ParseInt] will be used to parse the value.
//
// For an unsigned integer, [strconv.ParseUint] will be used to parse the
// value.
//
// For a float, [strconv.ParseFloat] will be used to parse the value.
//
// For a [time.Time], [dateparser.Parse] will be used to parse the value.
//
// For a [big.Int], [big.Int.SetString] will be used to parse the value.
//
// For a [big.Float], [big.Float.SetString] will be used to parse the value.
//
// For a type that implements [FlagValueSetter], the `SetFromFlagValue` method
// will be called and the [FlagValue] will be passed.
func (flag FlagValue) As(ctx context.Context, target any) error {
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr {
		return InvalidKindError{
			Value: target,
			Kind:  value.Kind(),
		}
	}

	instantiated, err := newValue(ctx, flag, value.Elem())
	if err != nil {
		return err
	}
	value.Elem().Set(instantiated)
	return nil
}
