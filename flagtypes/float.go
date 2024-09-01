package flagtypes

import (
	"context"
	"strconv"
	"strings"

	"impractical.co/clif"
)

// FloatParser is a [clif.FlagParser] implementation that can parse float64
// values.
type FloatParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [BasicFlag].
//
// The Value will be set to the result of [strconv.ParseFloat] for RawValue,
// assuming a 64 bit float.
func (FloatParser) Parse(_ context.Context, name, value string, _ clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	return BasicFlag[float64]{
		Name:     name,
		RawValue: value,
		Value:    parsed,
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// float flag.
func (FloatParser) FlagType() string {
	return "float"
}

// FloatListParser is a [clif.FlagParser] implementation that can parse values
// representing lists of floats, either specified as a comma-separated list or
// by specifying the flag multiple times.
//
// The results will be returned as a [ListFlag[float64]].
type FloatListParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [ListFlag[float64]]. The actual conversion is done by the
// [FloatParser.Parse] method.
//
// The RawValue will always use the comma-separated representation of the list,
// as there's no meaningful way to represent each flag usage.
func (FloatListParser) Parse(ctx context.Context, name, value string, prior clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	var list ListFlag[float64]
	if prior != nil {
		asserted, ok := prior.(ListFlag[float64])
		if !ok {
			return nil, UnexpectedFlagPriorTypeError{
				Name:     name,
				Expected: list,
				Got:      prior,
			}
		}
		list = asserted
	}
	basicVal, err := FloatParser{}.Parse(ctx, name, value, nil)
	if err != nil {
		return nil, err
	}
	floatFlag, ok := basicVal.(BasicFlag[float64])
	if !ok {
		return nil, UnexpectedFlagValueTypeError{
			Name:     name,
			Expected: BasicFlag[float64]{},
			Got:      basicVal,
		}
	}
	raw := make([]string, 0, len(list.Value))
	for _, val := range list.Value {
		raw = append(raw, strconv.FormatFloat(val, 'g', -1, 64))
	}
	return ListFlag[float64]{
		Name:     name,
		RawValue: strings.Join(append(raw, strconv.FormatFloat(floatFlag.Value, 'g', -1, 64)), ", "),
		Value:    append(list.Value, floatFlag.Value),
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// []float flag.
func (FloatListParser) FlagType() string {
	return "[]float"
}
