package clif

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestNewValueFromTime_success(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    FlagValue
		expected time.Time
	}
	testCases := map[string]testCase{
		"6 July 2020": {input: FlagValue{HasValue: true, Raw: "6 July 2020"}, expected: time.Date(2020, time.July, 6, 0, 0, 0, 0, time.UTC)},
		"07/06/2020":  {input: FlagValue{HasValue: true, Raw: "07/06/2020"}, expected: time.Date(2020, time.July, 6, 0, 0, 0, 0, time.UTC)},
		"07/06/20":    {input: FlagValue{HasValue: true, Raw: "07/06/20"}, expected: time.Date(2020, time.July, 6, 0, 0, 0, 0, time.UTC)},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target time.Time
			res, err := newValueFromTime(ctx, test.input, reflect.ValueOf(target))
			if err != nil {
				t.Fatalf("Unexpected error %s", err)
			}
			got, ok := res.Interface().(time.Time)
			if !ok {
				t.Fatalf("Expected result to be time.Time, was %T", res.Interface())
			}
			if !got.Equal(test.expected) {
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestNewValueFromTime_error(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":        "",
		"mixedCase":    "trUE",
		"invalidValue": "yes",
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target bool
			_, err := newValueFromTime(ctx, FlagValue{HasValue: true, Raw: input}, reflect.ValueOf(target))
			if err == nil {
				t.Fatal("Expected error, got none")
			}
		})
	}
}
