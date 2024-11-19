package clif

import (
	"context"
	"math/big"
	"reflect"
	"strconv"
)

func newValueFromInt(_ context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	var bitSize int
	switch target.Kind() { //nolint:exhaustive // we have a default assigned for a reason
	case reflect.Int:
		bitSize = 0
	case reflect.Int8:
		bitSize = 8
	case reflect.Int16:
		bitSize = 16
	case reflect.Int32:
		bitSize = 32
	case reflect.Int64:
		bitSize = 64
	default:
		return target, InvalidKindError{
			Value: target.Interface(),
			Kind:  target.Kind(),
		}
	}
	value, err := strconv.ParseInt(flag.Raw, 0, bitSize)
	if err != nil {
		return target, err
	}
	var result any
	switch target.Kind() { //nolint:exhaustive // we have a default assigned for a reason
	case reflect.Int:
		result = int(value)
	case reflect.Int8:
		result = int8(value) //nolint:gosec // we parsed it at the right size, we would have a strconv.ErrRange if there was going to be an overflow
	case reflect.Int16:
		result = int16(value) //nolint:gosec // we parsed it at the right size, we would have a strconv.ErrRange if there was going to be an overflow
	case reflect.Int32:
		result = int32(value) //nolint:gosec // we parsed it at the right size, we would have a strconv.ErrRange if there was going to be an overflow
	case reflect.Int64:
		result = value
	default:
		return target, InvalidKindError{
			Value: target.Interface(),
			Kind:  target.Kind(),
		}
	}
	return reflect.ValueOf(result).Convert(target.Type()), nil
}

func newValueFromUint(_ context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	var bitSize int
	switch target.Kind() { //nolint:exhaustive // we have a default assigned for a reason
	case reflect.Uint:
		bitSize = 0
	case reflect.Uint8:
		bitSize = 8
	case reflect.Uint16:
		bitSize = 16
	case reflect.Uint32:
		bitSize = 32
	case reflect.Uint64:
		bitSize = 64
	default:
		return target, InvalidKindError{
			Value: target.Interface(),
			Kind:  target.Kind(),
		}
	}
	value, err := strconv.ParseUint(flag.Raw, 0, bitSize)
	if err != nil {
		return target, err
	}
	var result any
	switch target.Kind() { //nolint:exhaustive // we have a default assigned for a reason
	case reflect.Uint:
		result = uint(value)
	case reflect.Uint8:
		result = uint8(value) //nolint:gosec // we parsed it at the right size, we would have a strconv.ErrRange if there was going to be an overflow
	case reflect.Uint16:
		result = uint16(value) //nolint:gosec // we parsed it at the right size, we would have a strconv.ErrRange if there was going to be an overflow
	case reflect.Uint32:
		result = uint32(value) //nolint:gosec // we parsed it at the right size, we would have a strconv.ErrRange if there was going to be an overflow
	case reflect.Uint64:
		result = value
	default:
		return target, InvalidKindError{
			Value: target.Interface(),
			Kind:  target.Kind(),
		}
	}
	return reflect.ValueOf(result).Convert(target.Type()), nil
}

func newValueFromFloat(_ context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	var bitSize int
	switch target.Kind() { //nolint:exhaustive // we have a default assigned for a reason
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	default:
		return target, InvalidKindError{
			Value: target.Interface(),
			Kind:  target.Kind(),
		}
	}
	value, err := strconv.ParseFloat(flag.Raw, bitSize)
	if err != nil {
		return target, err
	}
	var result any
	switch target.Kind() { //nolint:exhaustive // we have a default assigned for a reason
	case reflect.Float32:
		result = float32(value)
	case reflect.Float64:
		result = float64(value)
	default:
		return target, InvalidKindError{
			Value: target.Interface(),
			Kind:  target.Kind(),
		}
	}
	return reflect.ValueOf(result).Convert(target.Type()), nil
}

func newValueFromBigInt(_ context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	val := big.NewInt(0)
	val, ok := val.SetString(flag.Raw, 0)
	if !ok {
		return target, InvalidConversionError{
			Source: flag,
			Target: target,
		}
	}
	return reflect.ValueOf(val).Convert(target.Type()), nil
}

func newValueFromBigFloat(_ context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	val := big.NewFloat(0)
	val, ok := val.SetString(flag.Raw)
	if !ok {
		return target, InvalidConversionError{
			Source: flag,
			Target: target,
		}
	}
	return reflect.ValueOf(val).Convert(target.Type()), nil
}
