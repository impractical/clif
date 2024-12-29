package clif

import (
	"context"
	"reflect"
	"strconv"
)

func newValueFromBoolean(_ context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	value := true
	if flag.HasValue {
		parsed, err := strconv.ParseBool(flag.Raw)
		if err != nil {
			return target, err
		}
		value = parsed
	}
	return reflect.ValueOf(value).Convert(target.Type()), nil
}
