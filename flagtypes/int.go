package flagtypes

import (
	"context"
	"strconv"
	"strings"

	"impractical.co/clif"
)

// IntParser is a [clif.FlagParser] implementation that can parse int64 values.
type IntParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [BasicFlag].
//
// The Value will be set to the result of [strconv.ParseInt] for RawValue,
// assuming base 10 and a 64 bit integer.
func (IntParser) Parse(_ context.Context, name, value string, _ clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	return BasicFlag[int64]{
		Name:     name,
		RawValue: value,
		Value:    parsed,
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a int
// flag.
func (IntParser) FlagType() string {
	return "int"
}

// IntListParser is a [clif.FlagParser] implementation that can parse values
// representing lists of ints, either specified as a comma-separated list or by
// specifying the flag multiple times.
//
// The results will be returned as a [ListFlag[int64]].
type IntListParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [ListFlag[int64]]. The actual conversion is done by the
// [IntParser.Parse] method.
//
// The RawValue will always use the comma-separated representation of the list,
// as there's no meaningful way to represent each flag usage.
func (IntListParser) Parse(ctx context.Context, name, value string, prior clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	var list ListFlag[int64]
	if prior != nil {
		asserted, ok := prior.(ListFlag[int64])
		if !ok {
			return nil, UnexpectedFlagPriorTypeError{
				Name:     name,
				Expected: list,
				Got:      prior,
			}
		}
		list = asserted
	}
	basicVal, err := IntParser{}.Parse(ctx, name, value, nil)
	if err != nil {
		return nil, err
	}
	intFlag, ok := basicVal.(BasicFlag[int64])
	if !ok {
		return nil, UnexpectedFlagValueTypeError{
			Name:     name,
			Expected: BasicFlag[int64]{},
			Got:      basicVal,
		}
	}
	raw := make([]string, 0, len(list.Value))
	for _, val := range list.Value {
		raw = append(raw, strconv.FormatInt(val, 10))
	}
	return ListFlag[int64]{
		Name:     name,
		RawValue: strings.Join(append(raw, strconv.FormatInt(intFlag.Value, 10)), ", "),
		Value:    append(list.Value, intFlag.Value),
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// []int flag.
func (IntListParser) FlagType() string {
	return "[]int"
}
