package flagtypes

import (
	"context"
	"strconv"
	"strings"

	"impractical.co/clif"
)

// BoolParser is a [clif.FlagParser] implementation that can parse boolean
// values.
type BoolParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [BasicFlag].
//
// If the value is empty, the flag will be set to "true". Otherwise, the flag
// will be set to the [strconv.ParseBool] result for the value.
func (BoolParser) Parse(_ context.Context, name, value string, _ clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	// if we only have the flag name with no value, assume true
	val := true

	// if we have a value, use the value
	if value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		val = parsed
	}
	return BasicFlag[bool]{
		Name:     name,
		RawValue: value,
		Value:    val,
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a bool
// flag.
func (BoolParser) FlagType() string {
	return "bool"
}

// BoolListParser is a [clif.FlagParser] implementation that can parse values
// representing lists of bools, either specified as a comma-separated list or
// by specifying the flag multiple times.
//
// The results will be returned as a [ListFlag][bool].
type BoolListParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [ListFlag][bool]. The actual conversion is done by the
// [BoolParser.Parse] method.
//
// The RawValue will always use the comma-separated representation of the list,
// as there's no meaningful way to represent each flag usage.
func (BoolListParser) Parse(ctx context.Context, name, value string, prior clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	var list ListFlag[bool]
	if prior != nil {
		asserted, ok := prior.(ListFlag[bool])
		if !ok {
			return nil, UnexpectedFlagPriorTypeError{
				Name:     name,
				Expected: list,
				Got:      prior,
			}
		}
		list = asserted
	}
	basicVal, err := BoolParser{}.Parse(ctx, name, value, nil)
	if err != nil {
		return nil, err
	}
	boolFlag, ok := basicVal.(BasicFlag[bool])
	if !ok {
		return nil, UnexpectedFlagValueTypeError{
			Name:     name,
			Expected: BasicFlag[bool]{},
			Got:      basicVal,
		}
	}
	raw := make([]string, 0, len(list.Value))
	for _, val := range list.Value {
		raw = append(raw, strconv.FormatBool(val))
	}
	return ListFlag[bool]{
		Name:     name,
		RawValue: strings.Join(append(raw, strconv.FormatBool(boolFlag.Value)), ", "),
		Value:    append(list.Value, boolFlag.Value),
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// []bool flag.
func (BoolListParser) FlagType() string {
	return "[]bool"
}
