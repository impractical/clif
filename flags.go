package clif

import (
	"context"
	"fmt"
)

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
	// Parse is called to turn a name and value string into a Flag. The
	// prior value is passed to support flags that can be passed multiple
	// times to specify multiple values.
	Parse(ctx context.Context, name, value string, prior Flag) (Flag, error)

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
