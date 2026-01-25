package clif

import (
	"context"
	"math/big"
	"reflect"
	"time"
)

func newValue(ctx context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	if target.Type().Implements(reflect.TypeOf((*FlagValueSetter)(nil)).Elem()) {
		return newValueFromValueSetter(ctx, flag, target)
	}

	// we want to consider *big.Float and *big.Int to be numbers, not
	// pointers
	if target.Type() == reflect.TypeOf(big.NewFloat(0)) {
		return newValueFromBigFloat(ctx, flag, target)
	}
	if target.Type() == reflect.TypeOf(big.NewInt(0)) {
		return newValueFromBigInt(ctx, flag, target)
	}

	// we want to handle time.Time values even if they're structs
	if target.Type() == reflect.TypeOf(time.Time{}) {
		return newValueFromTime(ctx, flag, target)
	}

	switch target.Kind() { //nolint:exhaustive,nolintlint // we have a default assigned for a reason
	case reflect.Bool:
		return newValueFromBoolean(ctx, flag, target)
	case reflect.String:
		return reflect.ValueOf(flag.Raw).Convert(target.Type()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		return newValueFromInt(ctx, flag, target)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return newValueFromUint(ctx, flag, target)
	case reflect.Float32, reflect.Float64:
		return newValueFromFloat(ctx, flag, target)
	case reflect.Ptr:
		return newValueFromPointer(ctx, flag, target)
	case reflect.Struct, reflect.Slice, reflect.Map:
		return target, InvalidConversionError{
			Source: flag,
			Target: target,
		}
	default:
		return target, InvalidConversionError{
			Source: flag,
			Target: target,
		}
	}
}
