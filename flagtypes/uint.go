package flagtypes

import (
	"context"
	"strconv"
	"strings"

	"impractical.co/clif"
)

// UintParser is a [clif.FlagParser] implementation that can parse uint64
// values.
type UintParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [BasicFlag].
//
// The Value will be set to the result of [strconv.ParseUint] for RawValue,
// assuming base 10 and a 64 bit integer.
func (UintParser) Parse(_ context.Context, name, value string, _ clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, err
	}
	return BasicFlag[uint64]{
		Name:     name,
		RawValue: value,
		Value:    parsed,
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a uint
// flag.
func (UintParser) FlagType() string {
	return "uint"
}

// UintListParser is a [clif.FlagParser] implementation that can parse values
// representing lists of uints, either specified as a comma-separated list or
// by specifying the flag multiple times.
//
// The results will be returned as a [ListFlag[uint64]].
type UintListParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [ListFlag[uint64]]. The actual conversion is done by the
// [UintParser.Parse] method.
//
// The RawValue will always use the comma-separated representation of the list,
// as there's no meaningful way to represent each flag usage.
func (UintListParser) Parse(ctx context.Context, name, value string, prior clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	var list ListFlag[uint64]
	if prior != nil {
		asserted, ok := prior.(ListFlag[uint64])
		if !ok {
			return nil, UnexpectedFlagPriorTypeError{
				Name:     name,
				Expected: list,
				Got:      prior,
			}
		}
		list = asserted
	}
	basicVal, err := UintParser{}.Parse(ctx, name, value, nil)
	if err != nil {
		return nil, err
	}
	uintFlag, ok := basicVal.(BasicFlag[uint64])
	if !ok {
		return nil, UnexpectedFlagValueTypeError{
			Name:     name,
			Expected: BasicFlag[uint64]{},
			Got:      basicVal,
		}
	}
	raw := make([]string, 0, len(list.Value))
	for _, val := range list.Value {
		raw = append(raw, strconv.FormatUint(val, 10))
	}
	return ListFlag[uint64]{
		Name:     name,
		RawValue: strings.Join(append(raw, strconv.FormatUint(uintFlag.Value, 10)), ", "),
		Value:    append(list.Value, uintFlag.Value),
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// []uint flag.
func (UintListParser) FlagType() string {
	return "[]uint"
}
