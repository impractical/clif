package clif

import (
	"context"
	"reflect"
	"time"

	"github.com/markusmobius/go-dateparser"
)

func newValueFromTime(_ context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	value := time.Time{}
	if flag.HasValue {
		parsed, err := dateparser.Parse(nil, flag.Raw)
		if err != nil {
			return target, err
		}
		value = parsed.Time
	}
	return reflect.ValueOf(value).Convert(target.Type()), nil
}
