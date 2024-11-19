package clif

import (
	"fmt"
	"strings"
)

// UnknownFlagNameError is returned when an argument uses flag syntax, starting
// with a --, but doesn't match a flag configured for that [Command]. The
// underlying string will be the flag name, without leading --.
type UnknownFlagNameError string

func (err UnknownFlagNameError) Error() string {
	return fmt.Sprintf("unexpected flag %q", string(err))
}

// DuplicateFlagNameError is returned when multiple Commands use the same flag
// name, and it would be ambiguous which [Command] the flag applies to. The
// underlying string will be the flag name, without leading --.
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

// FlagValueWithoutFlagKeyError is returned when the parser finds a token that
// it believes can only be a flag value, but it's not preceded by a flag key.
// This should never happen.
type FlagValueWithoutFlagKeyError struct {
	Value string
}

func (err FlagValueWithoutFlagKeyError) Error() string {
	return fmt.Sprintf("somehow the flag parser decided %q was a flag value, but it's not preceded by a flag key, so that doesn't make sense; this is a bug and should be reported", err.Value)
}

// TooManyFlagValuesError is returned when a flag only accepts a single value,
// but multiple values were specified in the invocation.
type TooManyFlagValuesError string

func (err TooManyFlagValuesError) Error() string {
	return fmt.Sprintf("flag %q only accepts one value, but was specified multiple times", string(err))
}

// ShortFlagIsNotToggleError is returned when a [FlagDef] has a Name (or a key
// in Aliases) that makes it a short flag (the key starts with only one -), but
// is not marked as a toggle flag. Short flags must be toggle flags.
type ShortFlagIsNotToggleError struct {
	// FlagKey is the Name property of the FlagDef that is misconfigured.
	FlagKey string

	// Path is the Command, and any parents, that the FlagDef is defined
	// on. The last element in the slice will be the Command the FlagDef is
	// defined on.
	Path []string

	// BecauseOfAlias is the entry of FlagDef.Aliases that requires the
	// flag to be a short flag. If unset, it means the FlagDef.Name makes
	// the flag a short flag, not an alias.
	BecauseOfAlias string
}

func (err ShortFlagIsNotToggleError) Error() string {
	var shortFlagBecauseOfAlias string
	if err.BecauseOfAlias != "" {
		shortFlagBecauseOfAlias = fmt.Sprintf(" (must be a short flag because of alias %q)", err.BecauseOfAlias)
	}
	return fmt.Sprintf("flag %q (defined on %q) cannot be a short flag without IsToggle set to true%s", err.FlagKey, strings.Join(err.Path, " "), shortFlagBecauseOfAlias)
}

// OnlyToggleWithoutIsToggleError is returned when a [FlagDef] has its
// OnlyToggle property set to true but its IsToggle property set to false.
// OnlyToggle indicates that a flag can *only* be used without a value, but
// IsToggle is what makes the value optional to begin with, so having
// OnlyToggle true and IsToggle false means no invocations can ever be valid.
// This is obviously nonsensical, so it's an invalid configuration.
type OnlyToggleWithoutIsToggleError struct {
	// FlagKey is the Name property of the FlagDef that is misconfigured.
	FlagKey string

	// Path is the Command, and any parents, that the FlagDef is defined
	// on. The last element in the slice will be the Command the FlagDef is
	// defined on.
	Path []string
}

func (err OnlyToggleWithoutIsToggleError) Error() string {
	return fmt.Sprintf("flag %q (defined on %q) cannot have OnlyToggle set to true unless IsToggle is also true", err.FlagKey, strings.Join(err.Path, " "))
}

// InvalidFlagKeyError is returned when a flag key has an invalid name. Flag
// keys must start with -- (or just -, for short flags that are only used as
// toggles) and can only contain letters, numbers, -, _, and :. Flag keys must
// start with a letter or number, and short flags must be exactly one
// character.
type InvalidFlagKeyError struct {
	// CommandPath indicates the command and parents, if any, that the flag
	// is defined on.
	CommandPath []string

	// FlagKey is the invalid value used as a flag key.
	FlagKey string
}

func (err InvalidFlagKeyError) Error() string {
	return fmt.Sprintf("%q (defined on %q) is not a valid flag key; flag keys must start with -- (or - for short flags), can only contain letters, numbers, -, _, and :, must start with a letter or number, and short flags must be exactly one character", err.FlagKey, strings.Join(err.CommandPath, " "))
}
