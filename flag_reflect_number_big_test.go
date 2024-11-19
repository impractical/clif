package clif

import (
	"context"
	"errors"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"
)

func TestNewValueFromBigInt_success(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    FlagValue
		expected *big.Int
		target   any
	}
	testCases := map[string]testCase{
		// *big.Int
		"*big.Int/0":       {input: FlagValue{Set: true, Raw: "0"}, expected: big.NewInt(0), target: big.NewInt(0)},
		"*big.Int/-1":      {input: FlagValue{Set: true, Raw: "-1"}, expected: big.NewInt(-1), target: big.NewInt(0)},
		"*big.Int/1":       {input: FlagValue{Set: true, Raw: "1"}, expected: big.NewInt(1), target: big.NewInt(0)},
		"*big.Int/max-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(math.MaxInt64, 10)}, expected: big.NewInt(math.MaxInt64), target: big.NewInt(0)},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			res, err := newValueFromBigInt(ctx, test.input, reflect.ValueOf(test.target))
			if err != nil {
				t.Fatalf("Unexpected error %s", err)
			}
			got, ok := res.Interface().(*big.Int)
			if !ok {
				t.Fatalf("Result was not a *big.Int, was %T", res.Interface())
			}
			if got.Cmp(test.expected) != 0 {
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestNewValueFromBigInt_error_syntax(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":        "",
		"invalidValue": "not a number",
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target *big.Int
			_, err := newValueFromBigInt(ctx, FlagValue{Set: true, Raw: input}, reflect.ValueOf(target))
			if err == nil {
				t.Fatal("Expected error, got none")
			}
			conversionErr := InvalidConversionError{}
			if !errors.As(err, &conversionErr) {
				t.Fatalf("Expected clif.InvalidConversionError, got %T: %v", err, err)
			}
			if !conversionErr.Source.Set || conversionErr.Source.Raw != input {
				t.Fatalf("Expected conversion error's source to be set with a value of %q, got set: %v value: %q", input, conversionErr.Source.Set, conversionErr.Source.Raw)
			}
			if !conversionErr.Target.Equal(reflect.ValueOf(target)) {
				t.Fatalf("Expected conversion error's target to be %v, got %v", target, conversionErr.Target.Interface())
			}
		})
	}
}

func TestNewValueFromBigFloat_success(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    FlagValue
		expected *big.Float
		target   any
	}
	testCases := map[string]testCase{
		// *big.Float
		"*big.Float/0":         {input: FlagValue{Set: true, Raw: "0"}, expected: big.NewFloat(0), target: big.NewFloat(0)},
		"*big.Float/-1":        {input: FlagValue{Set: true, Raw: "-1"}, expected: big.NewFloat(-1), target: big.NewFloat(0)},
		"*big.Float/1":         {input: FlagValue{Set: true, Raw: "1"}, expected: big.NewFloat(1), target: big.NewFloat(0)},
		"*big.Float/max-float": {input: FlagValue{Set: true, Raw: strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64)}, expected: big.NewFloat(math.MaxFloat64), target: big.NewFloat(0)},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			res, err := newValueFromBigFloat(ctx, test.input, reflect.ValueOf(test.target))
			if err != nil {
				t.Fatalf("Unexpected error %s", err)
			}
			got, ok := res.Interface().(*big.Float)
			if !ok {
				t.Fatalf("Expected result to be *big.Float, was %T", res.Interface())
			}
			if got.Cmp(test.expected) != 0 {
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestNewValueFromBigFloat_error_syntax(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":        "",
		"invalidValue": "not a number",
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target *big.Float
			_, err := newValueFromBigFloat(ctx, FlagValue{Set: true, Raw: input}, reflect.ValueOf(target))
			if err == nil {
				t.Fatal("Expected error, got none")
			}
			conversionErr := InvalidConversionError{}
			if !errors.As(err, &conversionErr) {
				t.Fatalf("Expected clif.InvalidConversionError, got %T: %v", err, err)
			}
			if !conversionErr.Source.Set || conversionErr.Source.Raw != input {
				t.Fatalf("Expected conversion error's source to be set with a value of %q, got set: %v value: %q", input, conversionErr.Source.Set, conversionErr.Source.Raw)
			}
			if !conversionErr.Target.Equal(reflect.ValueOf(target)) {
				t.Fatalf("Expected conversion error's target to be %v, got %v", target, conversionErr.Target.Interface())
			}
		})
	}
}
