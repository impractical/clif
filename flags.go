package clif

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// UnknownFlagNameError is returned when an argument uses flag syntax, starting
// with a --, but doesn't match a flag configured for that [Command]. The
// underlying string will be the flag name, without leading --.
type UnknownFlagNameError string

func (err UnknownFlagNameError) Error() string {
	return fmt.Sprintf("unexpected flag %q", string(err))
}

// DuplicateFlagNameError is returned when multiple [Command]s use the same
// flag name, and it would be ambiguous which [Command] the flag applies to.
// The underlying string will be the flag name, without leading --.
type DuplicateFlagNameError string

func (err DuplicateFlagNameError) Error() string {
	return fmt.Sprintf("duplicate definitions of flag %q", string(err))
}

// UnexpectedFlagValueError is returned when a flag was used with a value, but
// the flag doesn't accept values.
type UnexpectedFlagValueError struct {
	// Flag is the flag name, without leading --.
	Flag string

	// Value is the value that was passed to the flag.
	Value string
}

func (err UnexpectedFlagValueError) Error() string {
	return fmt.Sprintf("value %q set for flag %q that doesn't accept values", err.Value, err.Flag)
}

// MissingFlagValueError is returned when a flag was used without a value, but
// the flag requires a value. The underlying string is the name of the flag,
// without leading --.
type MissingFlagValueError string

func (err MissingFlagValueError) Error() string {
	return fmt.Sprintf("no value set for flag that requires value %q", string(err))
}

// FlagDef holds the definition of a flag.
type FlagDef struct {
	// Name is the name of the flag. It's what will be surfaced in
	// documentation and what the user will use when applying the flag to a
	// command. Names must be unique across all commands, or the parser
	// won't know which command to apply the flag to.
	Name string

	// Aliases holds any alternative names the flag should accept from the
	// user. Aliases are not surfaced in documentation, by default. Aliases
	// must be unique across all other aliases and names for all commands,
	// or the parser won't know which command to apply the flag to.
	Aliases []string

	// Description is a user-friendly description of what the flag does and
	// what it's for, to be presented as part of help output.
	Description string

	// ValueAccepted indicates whether or not the flag should allow a
	// value. If set to false, attempting to pass a value will surface an
	// error.
	ValueAccepted bool

	// OnlyAfterCommandName indicates whether the flag should only be
	// acceptable after the command name in the invocation, or if can
	// appear anywhere in the invocation. If set to true, passing the flag
	// before the subcommand it belongs to will return an error.
	OnlyAfterCommandName bool

	// Parser determines how the flag value should be parsed.
	Parser FlagParser
}

// FlagParser is an interface for parsing flag values. Implementing it allows
// the definition of new types of flags.
type FlagParser interface {
	// Parse is called to turn a name and value string into a Flag.
	Parse(ctx context.Context, name, value string) (Flag, error)

	// FlagType should return a user-friendly indication of the type of
	// input this flag type expects, like "string" or "int" or "timestamp".
	FlagType() string
}

// Flag is an interface that holds information about a flag at runtime.
// Applications will almost always want to type assert this to the flag type
// returned by the [FlagParser] in the [FlagDef] for that specific flag, to get
// a parsed version of the value.
type Flag interface {
	GetName() string
	GetRawValue() string
}

// listFlagDefs recursively returns the list of [FlagDef]s defined on the
// passed [parseable] and all its subcommands.
func listFlagDefs(command parseable, activeCommand bool) []FlagDef {
	var flags []FlagDef
	for _, flag := range command.flags() {
		if !flag.OnlyAfterCommandName || activeCommand {
			flags = append(flags, flag)
		}
	}
	for _, sub := range command.subcommands() {
		flags = append(flags, listFlagDefs(sub, false)...)
	}
	return flags
}

// BasicFlagConstraint describes the types that the [BasicFlag] [Flag]
// implementation supports.
type BasicFlagConstraint interface {
	~bool | ~[]bool |
		~string | ~[]string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 | ~[]int | ~[]int8 | ~[]int16 | ~[]int32 | ~[]int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~[]uint | ~[]uint8 | ~[]uint16 | ~[]uint32 | ~[]uint64 |
		~float32 | ~float64 | ~[]float32 | ~[]float64 |
		~complex64 | ~complex128 | ~[]complex64 | ~[]complex128 |
		time.Time | []time.Time
}

// BasicFlag implements [Flag] for a base set of builtin types, allowing out of the box functionality similar to the [flag] package.
type BasicFlag[FlagType BasicFlagConstraint] struct {
	// Name will be set to the name the flag was invoked with.
	Name string

	// RawValue will be set to the string the user passed.
	RawValue string

	// Value will be set to the value that RawValue parsed into.
	Value FlagType
}

// GetName fills the [Flag] interface and returns the name the flag was invoked
// with.
func (flag BasicFlag[FlagType]) GetName() string {
	return flag.Name
}

// GetRawValue fills the [Flag] interface and returns the string the user
// passed as the flag's value.
func (flag BasicFlag[FlagType]) GetRawValue() string {
	return flag.RawValue
}

// BoolParser is a [FlagParser] implementation that can parse boolean values.
type BoolParser struct{}

// Parse fills the [FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// If the value is empty, the flag will be set to "true". Otherwise, the flag
// will be set to the [strconv.ParseBool] result for the value.
func (BoolParser) Parse(_ context.Context, name string, value string) (Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
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

// FlagType fills the [FlagParser] interface and identifies this as a bool
// flag.
func (BoolParser) FlagType() string {
	return "bool"
}

// StringParser is a [FlagParser] implementation that can parse string values.
type StringParser struct{}

// Parse fills the [FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// The Value and RawValue will always match.
func (StringParser) Parse(_ context.Context, name string, value string) (Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
	return BasicFlag[string]{
		Name:     name,
		RawValue: value,
		Value:    value,
	}, nil
}

// FlagType fills the [FlagParser] interface and identifies this as a string
// flag.
func (StringParser) FlagType() string {
	return "string"
}

// IntParser is a [FlagParser] implementation that can parse int64 values.
type IntParser struct{}

// Parse fills the [FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// The Value will be set to the result of [strconv.ParseInt] for RawValue,
// assuming base 10 and a 64 bit integer.
func (IntParser) Parse(_ context.Context, name string, value string) (Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
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

// FlagType fills the [FlagParser] interface and identifies this as a int flag.
func (IntParser) FlagType() string {
	return "int"
}

// UintParser is a [FlagParser] implementation that can parse uint64 values.
type UintParser struct{}

// Parse fills the [FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// The Value will be set to the result of [strconv.ParseUint] for RawValue,
// assuming base 10 and a 64 bit integer.
func (UintParser) Parse(_ context.Context, name string, value string) (Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
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

// FlagType fills the [FlagParser] interface and identifies this as a uint
// flag.
func (UintParser) FlagType() string {
	return "uint"
}

// FloatParser is a [FlagParser] implementation that can parse float64 values.
type FloatParser struct{}

// Parse fills the [FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// The Value will be set to the result of [strconv.ParseFloat] for RawValue,
// assuming a 64 bit float.
func (FloatParser) Parse(_ context.Context, name string, value string) (Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
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

// FlagType fills the [FlagParser] interface and identifies this as a float
// flag.
func (FloatParser) FlagType() string {
	return "float"
}

// TimeParser is a [FlagParser] implementation that can parse [time.Time]
// values.
type TimeParser struct{}

// Parse fills the [FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// Value will be set to the [time.Time] represented by the RawValue. Only the
// [time.RFC3339Nano] format is supported at the moment.
func (TimeParser) Parse(_ context.Context, name string, value string) (Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
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

// FlagType fills the [FlagParser] interface and identifies this as a timestamp
// flag.
func (TimeParser) FlagType() string {
	return "timestamp"
}

// DurationParser is a [FlagParser] implementation that can parse
// [time.Duration] values.
type DurationParser struct{}

// Parse fills the [FlagParser] interface and converts a name and value into a
// [BasicFlag].
//
// Value will be set to the [time.Duration] returned by [time.ParseDuration]
// when passed the RawValue.
func (DurationParser) Parse(_ context.Context, name string, value string) (Flag, error) { //nolint:ireturn // FlagParser interface requires returning an interface
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

// FlagType fills the [FlagParser] interface and identifies this as a duration
// flag.
func (DurationParser) FlagType() string {
	return "duration"
}
