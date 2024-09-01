package flagtypes

import (
	"context"
	"strings"

	"impractical.co/clif"
)

// StringParser is a [clif.FlagParser] implementation that can parse string
// values.
type StringParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [BasicFlag].
//
// The Value and RawValue will always match.
func (StringParser) Parse(_ context.Context, name, value string, _ clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	return BasicFlag[string]{
		Name:     name,
		RawValue: value,
		Value:    value,
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// string flag.
func (StringParser) FlagType() string {
	return "string"
}

// StringListParser is a [clif.FlagParser] implementation that can parse values
// representing lists of strings, either specified as a comma-separated list or
// by specifying the flag multiple times.
//
// The results will be returned as a [ListFlag[string]].
type StringListParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [ListFlag[string]].
//
// The RawValue will always use the comma-separated representation of the list,
// as there's no meaningful way to represent each flag usage.
func (StringListParser) Parse(ctx context.Context, name, value string, prior clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	var list ListFlag[string]
	if prior != nil {
		asserted, ok := prior.(ListFlag[string])
		if !ok {
			return nil, UnexpectedFlagPriorTypeError{
				Name:     name,
				Expected: list,
				Got:      prior,
			}
		}
		list = asserted
	}
	basicVal, err := StringParser{}.Parse(ctx, name, value, nil)
	if err != nil {
		return nil, err
	}
	stringFlag, ok := basicVal.(BasicFlag[string])
	if !ok {
		return nil, UnexpectedFlagValueTypeError{
			Name:     name,
			Expected: BasicFlag[string]{},
			Got:      basicVal,
		}
	}
	return ListFlag[string]{
		Name:     name,
		RawValue: strings.Join(append(list.Value, stringFlag.Value), ", "),
		Value:    append(list.Value, stringFlag.Value),
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// []string flag.
func (StringListParser) FlagType() string {
	return "[]string"
}
