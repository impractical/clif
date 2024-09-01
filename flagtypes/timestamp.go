package flagtypes

import (
	"context"
	"strings"
	"time"

	"impractical.co/clif"
)

// TimeParser is a [clif.FlagParser] implementation that can parse [time.Time]
// values.
type TimeParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [BasicFlag].
//
// Value will be set to the [time.Time] represented by the RawValue. Only the
// [time.RFC3339Nano] format is supported at the moment.
func (TimeParser) Parse(_ context.Context, name, value string, _ clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, err
	}
	return BasicFlag[time.Time]{
		Name:     name,
		RawValue: value,
		Value:    parsed,
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// timestamp flag.
func (TimeParser) FlagType() string {
	return "timestamp"
}

// TimeListParser is a [clif.FlagParser] implementation that can parse values
// representing lists of timestamps, either specified as a comma-separated list
// or by specifying the flag multiple times.
//
// The results will be returned as a [ListFlag][time.Time].
type TimeListParser struct{}

// Parse fills the [clif.FlagParser] interface and converts a name and value
// into a [ListFlag][time.Time]. The actual conversion is done by the
// [TimeParser.Parse] method.
//
// The RawValue will always use the comma-separated representation of the list,
// as there's no meaningful way to represent each flag usage.
func (TimeListParser) Parse(ctx context.Context, name, value string, prior clif.Flag) (clif.Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	var list ListFlag[time.Time]
	if prior != nil {
		asserted, ok := prior.(ListFlag[time.Time])
		if !ok {
			return nil, UnexpectedFlagPriorTypeError{
				Name:     name,
				Expected: list,
				Got:      prior,
			}
		}
		list = asserted
	}
	basicVal, err := TimeParser{}.Parse(ctx, name, value, nil)
	if err != nil {
		return nil, err
	}
	timeFlag, ok := basicVal.(BasicFlag[time.Time])
	if !ok {
		return nil, UnexpectedFlagValueTypeError{
			Name:     name,
			Expected: BasicFlag[time.Time]{},
			Got:      basicVal,
		}
	}
	raw := make([]string, 0, len(list.Value))
	for _, val := range list.Value {
		raw = append(raw, val.Format(time.RFC3339Nano))
	}
	return ListFlag[time.Time]{
		Name:     name,
		RawValue: strings.Join(append(raw, timeFlag.Value.Format(time.RFC3339Nano)), ", "),
		Value:    append(list.Value, timeFlag.Value),
	}, nil
}

// FlagType fills the [clif.FlagParser] interface and identifies this as a
// []timestamp flag.
func (TimeListParser) FlagType() string {
	return "[]timestamp"
}
