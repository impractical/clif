package flagtypes

import (
	"context"
	"strings"
	"time"

	"impractical.co/clif"
)

// DurationParser is a [clif.FlagParser] implementation that can parse
// [time.Duration] values.
type DurationParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// Value will be set to the [time.Duration] returned by [time.ParseDuration]
// when passed the RawValue.
func (DurationParser) Parse(_ context.Context, name, value string, _ clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return nil, err
	}
	return BasicFlag[time.Duration]{
		Name:     name,
		RawValue: value,
		Value:    parsed,
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a duration
// flag.
func (DurationParser) FlagType() string {
	return "duration"
}

// DurationListParser is a [clif.FlagParser] implementation that can parse values
// representing lists of durations, either specified as a comma-separated list
// or by specifying the flag multiple times.
//
// The results will be returned as a ListFlag[[]time.Duration].
type DurationListParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value into a
// [ListFlag][[][time.Duration]]. The actual conversion is done by the
// [DurationParser.Parse] method.
//
// The RawValue will always use the comma-separated representation of the list,
// as there's no meaningful way to represent each flag usage.
func (DurationListParser) Parse(ctx context.Context, name, value string, prior clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	var list ListFlag[time.Duration]
	if prior != nil {
		asserted, ok := prior.(ListFlag[time.Duration])
		if !ok {
			return nil, UnexpectedFlagPriorTypeError{
				Name:     name,
				Expected: list,
				Got:      prior,
			}
		}
		list = asserted
	}
	basicVal, err := DurationParser{}.Parse(ctx, name, value, nil)
	if err != nil {
		return nil, err
	}
	durationFlag, ok := basicVal.(BasicFlag[time.Duration])
	if !ok {
		return nil, UnexpectedFlagValueTypeError{
			Name:     name,
			Expected: BasicFlag[time.Duration]{},
			Got:      basicVal,
		}
	}
	raw := make([]string, 0, len(list.Value))
	for _, val := range list.Value {
		raw = append(raw, val.String())
	}
	return ListFlag[time.Duration]{
		Name:     name,
		RawValue: strings.Join(append(raw, durationFlag.Value.String()), ", "),
		Value:    append(list.Value, durationFlag.Value),
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a int
// flag.
func (DurationListParser) FlagType() string {
	return "[]duration"
}
