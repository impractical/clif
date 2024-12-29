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

func TestNewValueFromFloat_success(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    FlagValue
		expected float64
		target   any
	}
	testCases := map[string]testCase{
		// float32
		"float32/0":         {input: FlagValue{HasValue: true, Raw: "0"}, expected: float64(0), target: float32(0)},
		"float32/-1":        {input: FlagValue{HasValue: true, Raw: "-1"}, expected: float64(-1), target: float32(0)},
		"float32/1":         {input: FlagValue{HasValue: true, Raw: "1"}, expected: float64(1), target: float32(0)},
		"float32/max-float": {input: FlagValue{HasValue: true, Raw: strconv.FormatFloat(float64(math.MaxFloat32), 'f', -1, 32)}, expected: math.MaxFloat32, target: float32(0)},

		// float64
		"float64/0":         {input: FlagValue{HasValue: true, Raw: "0"}, expected: float64(0), target: float64(0)},
		"float64/-1":        {input: FlagValue{HasValue: true, Raw: "-1"}, expected: float64(-1), target: float64(0)},
		"float64/1":         {input: FlagValue{HasValue: true, Raw: "1"}, expected: float64(1), target: float64(0)},
		"float64/max-float": {input: FlagValue{HasValue: true, Raw: strconv.FormatFloat(float64(math.MaxFloat64), 'f', -1, 64)}, expected: math.MaxFloat64, target: float64(0)},
		"float64/inf":       {input: FlagValue{HasValue: true, Raw: "inf"}, expected: math.Inf(1), target: float64(0)},
		"float64/infinity":  {input: FlagValue{HasValue: true, Raw: "infinity"}, expected: math.Inf(1), target: float64(0)},
		"float64/-inf":      {input: FlagValue{HasValue: true, Raw: "-inf"}, expected: math.Inf(-1), target: float64(0)},
		"float64/-infinity": {input: FlagValue{HasValue: true, Raw: "-infinity"}, expected: math.Inf(-1), target: float64(0)},
		"float64/+inf":      {input: FlagValue{HasValue: true, Raw: "+inf"}, expected: math.Inf(1), target: float64(0)},
		"float64/+infinity": {input: FlagValue{HasValue: true, Raw: "+infinity"}, expected: math.Inf(1), target: float64(0)},
		"float64/nan":       {input: FlagValue{HasValue: true, Raw: "NaN"}, expected: math.NaN(), target: float64(0)},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			res, err := newValueFromFloat(ctx, test.input, reflect.ValueOf(test.target))
			if err != nil {
				t.Fatalf("Unexpected error %s", err)
			}
			got := res.Float()
			if got != test.expected {
				if math.IsNaN(test.expected) && math.IsNaN(got) {
					return
				}
				if math.IsInf(test.expected, -1) && math.IsInf(got, -1) {
					return
				}
				if math.IsInf(test.expected, 1) && math.IsInf(got, 1) {
					return
				}
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestNewValueFromFloat_error_syntax(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":        "",
		"invalidValue": "not a number",
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target float64
			_, err := newValueFromFloat(ctx, FlagValue{HasValue: true, Raw: input}, reflect.ValueOf(target))
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

func TestNewValueFromFloat_error_overflow(t *testing.T) {
	t.Parallel()

	type testCase struct {
		target any
		value  *big.Float
	}

	testCases := map[string]testCase{
		"float32": {
			value:  big.NewFloat(0).Add(big.NewFloat(float64(math.MaxFloat32)), big.NewFloat(math.MaxFloat32/2)),
			target: float32(0),
		},
		"float64": {
			value:  big.NewFloat(0).Add(big.NewFloat(float64(math.MaxFloat64)), big.NewFloat(1000)),
			target: float64(0),
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			_, err := newValueFromFloat(ctx, FlagValue{HasValue: true, Raw: test.value.String()}, reflect.ValueOf(test.target))
			if err == nil {
				t.Fatal("Expected error, got none")
			}
			numError := &strconv.NumError{}
			if !errors.As(err, &numError) {
				t.Fatalf("Expected strconv.NumError, got %T: %v", err, err)
			}
			if !errors.Is(numError.Err, strconv.ErrRange) {
				t.Fatalf("Expected strconv.NumError to be reporting a strconv.ErrRange, got %v instead", numError.Err)
			}
		})
	}
}
