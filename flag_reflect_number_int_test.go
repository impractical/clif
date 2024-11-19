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

func TestNewValueFromInt_success(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    FlagValue
		expected int64
		target   any
	}
	testCases := map[string]testCase{
		// int
		"int/0":       {input: FlagValue{Set: true, Raw: "0"}, expected: int64(0), target: int(0)},
		"int/-1":      {input: FlagValue{Set: true, Raw: "-1"}, expected: int64(-1), target: int(0)},
		"int/1":       {input: FlagValue{Set: true, Raw: "1"}, expected: int64(1), target: int(0)},
		"int/max-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MaxInt), 10)}, expected: math.MaxInt, target: int(0)},
		"int/min-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MinInt), 10)}, expected: math.MinInt, target: int(0)},

		// int8
		"int8/0":       {input: FlagValue{Set: true, Raw: "0"}, expected: int64(0), target: int8(0)},
		"int8/-1":      {input: FlagValue{Set: true, Raw: "-1"}, expected: int64(-1), target: int8(0)},
		"int8/1":       {input: FlagValue{Set: true, Raw: "1"}, expected: int64(1), target: int8(0)},
		"int8/max-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MaxInt8), 10)}, expected: math.MaxInt8, target: int8(0)},
		"int8/min-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MinInt8), 10)}, expected: math.MinInt8, target: int8(0)},

		// int16
		"int16/0":       {input: FlagValue{Set: true, Raw: "0"}, expected: int64(0), target: int16(0)},
		"int16/-1":      {input: FlagValue{Set: true, Raw: "-1"}, expected: int64(-1), target: int16(0)},
		"int16/1":       {input: FlagValue{Set: true, Raw: "1"}, expected: int64(1), target: int16(0)},
		"int16/max-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MaxInt16), 10)}, expected: math.MaxInt16, target: int16(0)},
		"int16/min-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MinInt16), 10)}, expected: math.MinInt16, target: int16(0)},

		// int32
		"int32/0":       {input: FlagValue{Set: true, Raw: "0"}, expected: int64(0), target: int32(0)},
		"int32/-1":      {input: FlagValue{Set: true, Raw: "-1"}, expected: int64(-1), target: int32(0)},
		"int32/1":       {input: FlagValue{Set: true, Raw: "1"}, expected: int64(1), target: int32(0)},
		"int32/max-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MaxInt32), 10)}, expected: math.MaxInt32, target: int32(0)},
		"int32/min-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MinInt32), 10)}, expected: math.MinInt32, target: int32(0)},

		// int64
		"int64/0":       {input: FlagValue{Set: true, Raw: "0"}, expected: int64(0), target: int64(0)},
		"int64/-1":      {input: FlagValue{Set: true, Raw: "-1"}, expected: int64(-1), target: int64(0)},
		"int64/1":       {input: FlagValue{Set: true, Raw: "1"}, expected: int64(1), target: int64(0)},
		"int64/max-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MaxInt64), 10)}, expected: math.MaxInt64, target: int64(0)},
		"int64/min-int": {input: FlagValue{Set: true, Raw: strconv.FormatInt(int64(math.MinInt64), 10)}, expected: math.MinInt64, target: int64(0)},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			res, err := newValueFromInt(ctx, test.input, reflect.ValueOf(test.target))
			if err != nil {
				t.Fatalf("Unexpected error %s", err)
			}
			got := res.Int()
			if got != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestNewValueFromInt_error_syntax(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":        "",
		"invalidValue": "NaN",
		"float":        "12.34",
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target int
			_, err := newValueFromInt(ctx, FlagValue{Set: true, Raw: input}, reflect.ValueOf(target))
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

func TestNewValueFromInt_error_overflow(t *testing.T) {
	t.Parallel()

	type testCase struct {
		target any
		value  *big.Int
	}

	testCases := map[string]testCase{
		"int": {
			value:  big.NewInt(0).Add(big.NewInt(int64(math.MaxInt)), big.NewInt(1)),
			target: int(0),
		},
		"int8": {
			value:  big.NewInt(0).Add(big.NewInt(int64(math.MaxInt8)), big.NewInt(1)),
			target: int8(0),
		},
		"int16": {
			value:  big.NewInt(0).Add(big.NewInt(int64(math.MaxInt16)), big.NewInt(1)),
			target: int16(0),
		},
		"int32": {
			value:  big.NewInt(0).Add(big.NewInt(int64(math.MaxInt32)), big.NewInt(1)),
			target: int32(0),
		},
		"int64": {
			value:  big.NewInt(0).Add(big.NewInt(int64(math.MaxInt64)), big.NewInt(1)),
			target: int64(0),
		},
		"int-underflow": {
			value:  big.NewInt(0).Sub(big.NewInt(int64(math.MinInt)), big.NewInt(1)),
			target: int(0),
		},
		"int8-underflow": {
			value:  big.NewInt(0).Sub(big.NewInt(int64(math.MinInt8)), big.NewInt(1)),
			target: int8(0),
		},
		"int16-underflow": {
			value:  big.NewInt(0).Sub(big.NewInt(int64(math.MinInt16)), big.NewInt(1)),
			target: int16(0),
		},
		"int32-underflow": {
			value:  big.NewInt(0).Sub(big.NewInt(int64(math.MinInt32)), big.NewInt(1)),
			target: int32(0),
		},
		"int64-underflow": {
			value:  big.NewInt(0).Sub(big.NewInt(int64(math.MinInt64)), big.NewInt(1)),
			target: int64(0),
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			_, err := newValueFromInt(ctx, FlagValue{Set: true, Raw: test.value.String()}, reflect.ValueOf(test.target))
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
