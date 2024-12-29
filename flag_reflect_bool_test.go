package clif

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"
)

func TestNewValueFromBoolean_success(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    FlagValue
		expected bool
	}
	testCases := map[string]testCase{
		"true":   {input: FlagValue{HasValue: true, Raw: "true"}, expected: true},
		"TRUE":   {input: FlagValue{HasValue: true, Raw: "TRUE"}, expected: true},
		"t":      {input: FlagValue{HasValue: true, Raw: "t"}, expected: true},
		"1":      {input: FlagValue{HasValue: true, Raw: "1"}, expected: true},
		"false":  {input: FlagValue{HasValue: true, Raw: "false"}, expected: false},
		"FALSE":  {input: FlagValue{HasValue: true, Raw: "FALSE"}, expected: false},
		"f":      {input: FlagValue{HasValue: true, Raw: "f"}, expected: false},
		"0":      {input: FlagValue{HasValue: true, Raw: "0"}, expected: false},
		"toggle": {input: FlagValue{HasValue: false, Raw: ""}, expected: true},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target bool
			res, err := newValueFromBoolean(ctx, test.input, reflect.ValueOf(target))
			if err != nil {
				t.Fatalf("Unexpected error %s", err)
			}
			got := res.Bool()
			if got != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestNewValueFromBoolean_error(t *testing.T) {
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
			_, err := newValueFromBoolean(ctx, FlagValue{HasValue: true, Raw: input}, reflect.ValueOf(target))
			if err == nil {
				t.Fatal("Expected error, got none")
			}
			numError := &strconv.NumError{}
			if !errors.As(err, &numError) {
				t.Fatalf("Expected strconv.NumError, got %T: %v", err, err)
			}
			if !errors.Is(numError.Err, strconv.ErrSyntax) {
				t.Fatalf("Expected strconv.NumError to be reporting a strconv.ErrSyntax, got %v instead", numError.Err)
			}
		})
	}
}
