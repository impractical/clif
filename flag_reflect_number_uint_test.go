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

func TestNewValueFromUint_success(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    FlagValue
		expected uint64
		target   any
	}
	testCases := map[string]testCase{
		// uint
		"uint/0":        {input: FlagValue{HasValue: true, Raw: "0"}, expected: uint64(0), target: uint(0)},
		"uint/1":        {input: FlagValue{HasValue: true, Raw: "1"}, expected: uint64(1), target: uint(0)},
		"uint/max-uint": {input: FlagValue{HasValue: true, Raw: strconv.FormatUint(uint64(math.MaxUint), 10)}, expected: math.MaxUint, target: uint(0)},

		// uint8
		"uint8/0":        {input: FlagValue{HasValue: true, Raw: "0"}, expected: uint64(0), target: uint8(0)},
		"uint8/1":        {input: FlagValue{HasValue: true, Raw: "1"}, expected: uint64(1), target: uint8(0)},
		"uint8/max-uint": {input: FlagValue{HasValue: true, Raw: strconv.FormatUint(uint64(math.MaxUint8), 10)}, expected: math.MaxUint8, target: uint8(0)},

		// uint16
		"uint16/0":        {input: FlagValue{HasValue: true, Raw: "0"}, expected: uint64(0), target: uint16(0)},
		"uint16/1":        {input: FlagValue{HasValue: true, Raw: "1"}, expected: uint64(1), target: uint16(0)},
		"uint16/max-uint": {input: FlagValue{HasValue: true, Raw: strconv.FormatUint(uint64(math.MaxUint16), 10)}, expected: math.MaxUint16, target: uint16(0)},

		// uint32
		"uint32/0":        {input: FlagValue{HasValue: true, Raw: "0"}, expected: uint64(0), target: uint32(0)},
		"uint32/1":        {input: FlagValue{HasValue: true, Raw: "1"}, expected: uint64(1), target: uint32(0)},
		"uint32/max-uint": {input: FlagValue{HasValue: true, Raw: strconv.FormatUint(uint64(math.MaxUint32), 10)}, expected: math.MaxUint32, target: uint32(0)},

		// uint64
		"uint64/0":        {input: FlagValue{HasValue: true, Raw: "0"}, expected: uint64(0), target: uint64(0)},
		"uint64/1":        {input: FlagValue{HasValue: true, Raw: "1"}, expected: uint64(1), target: uint64(0)},
		"uint64/max-uint": {input: FlagValue{HasValue: true, Raw: strconv.FormatUint(uint64(math.MaxUint64), 10)}, expected: math.MaxUint64, target: uint64(0)},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			res, err := newValueFromUint(ctx, test.input, reflect.ValueOf(test.target))
			if err != nil {
				t.Fatalf("Unexpected error %s", err)
			}
			got := res.Uint()
			if got != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestNewValueFromUint_error_syntax(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":        "",
		"invalidValue": "NaN",
		"float":        "12.34",
		"negative":     "-1",
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var target uint
			_, err := newValueFromUint(ctx, FlagValue{HasValue: true, Raw: input}, reflect.ValueOf(target))
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

func TestNewValueFromUint_error_overflow(t *testing.T) {
	t.Parallel()

	type testCase struct {
		target any
		value  *big.Int
	}

	testCases := map[string]testCase{
		"uint": {
			value:  big.NewInt(0).Add(big.NewInt(0).SetUint64(uint64(math.MaxUint)), big.NewInt(1)),
			target: uint(0),
		},
		"uint8": {
			value:  big.NewInt(0).Add(big.NewInt(0).SetUint64(uint64(math.MaxUint8)), big.NewInt(1)),
			target: uint8(0),
		},
		"uint16": {
			value:  big.NewInt(0).Add(big.NewInt(0).SetUint64(uint64(math.MaxUint16)), big.NewInt(1)),
			target: uint16(0),
		},
		"uint32": {
			value:  big.NewInt(0).Add(big.NewInt(0).SetUint64(uint64(math.MaxUint32)), big.NewInt(1)),
			target: uint32(0),
		},
		"uint64": {
			value:  big.NewInt(0).Add(big.NewInt(0).SetUint64(uint64(math.MaxUint64)), big.NewInt(1)),
			target: uint64(0),
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			_, err := newValueFromUint(ctx, FlagValue{HasValue: true, Raw: test.value.String()}, reflect.ValueOf(test.target))
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
