package clif

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"strings"
)

// UnexpectedCommandArgError is returned when a command that wasn't expecting
// an argument gets one.
type UnexpectedCommandArgError string

func (err UnexpectedCommandArgError) Error() string {
	return fmt.Sprintf("unexpected argument: %s", string(err))
}

// Command defines a command the user can run. Commands can have handlers, that
// get invoked when the command is run, and subcommands, which are other
// commands namespaced under their command. Commands with subcommands can still
// be invoked, and should still have a handler defined, even if it just prints
// out usage information on the subcommands.
type Command struct {
	// Name is the name of the command, what the user will type to prompt
	// its functionality.
	Name string

	// Aliases are acceptable variations on Name; they will be treated as
	// equivalent to Name, but will not be listed in the SubcommandsHelp
	// output.
	Aliases []string

	// Description is a short, one-line description of the command, used
	// when generating the SubcommandsHelp output.
	Description string

	// Hidden indicates whether a command should be included in
	// SubcommandsHelp output or not. If set to true, the command will be
	// omitted from SubcommandsHelp output.
	Hidden bool

	// Flags holds definitions for the flags, if any, that this command
	// accepts.
	Flags []FlagDef

	// Subcommands are the various subcommands, if any, that this command
	// accepts.
	Subcommands []Command

	// Handler is the HandlerBuilder executed when this Command is used.
	// The Handler will not be executed if a subcommand of this Command is
	// used.
	Handler HandlerBuilder

	// ArgsAccepted indicates whether free input is expected as part of
	// this command. If true, this command cannot have any subcommands.
	ArgsAccepted bool
}

// Validate determines whether a [Command] has a valid definition or not.
func (cmd Command) Validate(ctx context.Context, path []string, parentFlags map[string]struct{}) error {
	var errs error
	parentPath := path
	if len(path) > 0 {
		parentPath = path[:len(path)-1]
	}
	errs = errors.Join(errs, validateCommandName(cmd.Name, append(parentPath, cmd.Name), false))
	for _, alias := range cmd.Aliases {
		errs = errors.Join(errs, validateCommandName(alias, append(parentPath, alias), false))
	}
	if len(cmd.Subcommands) < 1 && cmd.Handler == nil {
		errs = errors.Join(errs, CommandMissingSubcommandsOrHandlerError{Path: path})
	}
	if len(cmd.Subcommands) > 0 && cmd.ArgsAccepted {
		errs = errors.Join(errs, CommandAcceptsArgumentsAndSubcommandsError{Path: path})
	}
	flagKeys := maps.Clone(parentFlags)
	for _, flagDef := range cmd.Flags {
		if _, ok := flagKeys[flagDef.Name]; ok {
			errs = errors.Join(errs, DuplicateFlagNameError(flagDef.Name))
		}
		flagKeys[flagDef.Name] = struct{}{}
		for _, alias := range flagDef.Aliases {
			if _, ok := flagKeys[alias]; ok {
				errs = errors.Join(errs, DuplicateFlagNameError(alias))
			}
			flagKeys[alias] = struct{}{}
		}
		errs = errors.Join(errs, flagDef.Validate(ctx, path))
	}
	subNames := map[string]struct{}{}
	for pos, sub := range cmd.Subcommands {
		if sub.Name == "" {
			errs = errors.Join(errs, CommandMissingNameError{
				Path: path,
				Pos:  pos,
			})
			continue
		}
		if _, ok := subNames[sub.Name]; ok {
			errs = errors.Join(errs, DuplicateCommandError{Path: path, Command: sub.Name})
		}
		for _, alias := range sub.Aliases {
			if alias == "" {
				errs = errors.Join(errs, CommandAliasEmptyError{Path: path, Command: sub.Name})
			} else if alias == sub.Name {
				errs = errors.Join(errs, CommandDuplicatesNameAsAliasError{Path: path, Command: sub.Name})
			} else if _, ok := subNames[alias]; ok {
				errs = errors.Join(errs, DuplicateCommandError{Path: path, Command: alias})
			}
		}
		errs = errors.Join(errs, sub.Validate(ctx, append(path, sub.Name), flagKeys))
	}
	return errs
}

var (
	commandNameRE = regexp.MustCompile(`^[a-zA-Z0-9-_:]+$`)
)

func validateCommandName(name string, path []string, alias bool) error {
	if name[0] == '-' {
		return InvalidCommandNameError{
			Name:  name,
			Path:  path,
			Alias: alias,
		}
	}
	if commandNameRE.MatchString(name) {
		return nil
	}
	return InvalidCommandNameError{
		Name:  name,
		Path:  path,
		Alias: alias,
	}
}

// CommandMissingNameError is returned when a [Command], either defined on
// [Application.Commands] or [Command.Subcommands], doesn't have its Name
// property set.
type CommandMissingNameError struct {
	// Pos is the position of the Command in Application.Commands or
	// Command.Subcommands without a Name property set. Because there's no
	// Name property, we have no other way to indicate which Command we're
	// talking about.
	Pos int

	// Path is the list of Commands that were traversed to get to the
	// Command that's missing a Name. An empty slice indicates the Command
	// is defined in Application.Commands. Otherwise, the Command is
	// defined in the Subcommands property of the last Command in the
	// slice.
	Path []string
}

func (err CommandMissingNameError) Error() string {
	commandOrSubcommand := "command"
	// this kind of error is returned from the parent validator, not the
	// command's validator, so the command won't be in the path yet
	if len(err.Path) > 0 {
		commandOrSubcommand = fmt.Sprintf("subcommand of %q", strings.Join(err.Path, " "))
	}
	return fmt.Sprintf("%s defined in position %d is missing a name", commandOrSubcommand, err.Pos)
}

// CommandMissingSubcommandsOrHandlerError is returned when a [Command] has no
// Subcommands or Handler set, meaning it can't do anything, which is invalid.
type CommandMissingSubcommandsOrHandlerError struct {
	// Path indicates that parents, if any, of the Command with no
	// Subcommands or Handler defined. The Command in question will be the
	// last entry in the slice.
	Path []string
}

func (err CommandMissingSubcommandsOrHandlerError) Error() string {
	return fmt.Sprintf("%q must define either subcommands or a handler", strings.Join(err.Path, " "))
}

// CommandAcceptsArgumentsAndSubcommandsError is returned when a [Command] is
// defined that accepts both arguments and subcommands, which is invalid. A
// [Command] may only accept one.
type CommandAcceptsArgumentsAndSubcommandsError struct {
	// Path indicates that parents, if any, of the Command that accepts
	// both arguments and subcommands. The Command in question will be the
	// last entry in the slice.
	Path []string
}

func (err CommandAcceptsArgumentsAndSubcommandsError) Error() string {
	return fmt.Sprintf("%q must accept either subcommands or arguments, it cannot accept both", strings.Join(err.Path, " "))
}

// DuplicateCommandError is returned when a [Command] is defined in the an
// [Application.Commands] or [Command.Subcommands] that already uses the
// [Command.Name] or one of the [Command.Aliases] for another [Command].
// [Command.Name] and [Command.Aliases] must be unique within their parent.
type DuplicateCommandError struct {
	// Path is the list of parents, if any, of the Commands that have
	// reused the same name or alias. If empty, this indicates that
	// Application.Commands are in conflict; otherwise, the last element in
	// the slice is the Command containing the conflicting Commands.
	Path []string

	// Command is the name that has been reused.
	Command string
}

func (err DuplicateCommandError) Error() string {
	commandOrSubcommand := "commands"
	if len(err.Path) > 0 {
		commandOrSubcommand = fmt.Sprintf("subcommands of %s", strings.Join(err.Path, " "))
	}
	return fmt.Sprintf("multiple %s defined using the name or alias %q", commandOrSubcommand, err.Command)
}

// CommandAliasEmptyError is returned when the [Command.Alias] property
// includes an empty string, which is invalid.
type CommandAliasEmptyError struct {
	// Path is the list of parents, if any, of the Command with an empty
	// string in the Aliases slice. If empty, this indicates the Command is
	// in Application.Commands.
	Path []string

	// Command is the name of the Command with an empty string in its
	// Aliases list.
	Command string
}

func (err CommandAliasEmptyError) Error() string {
	commandOrSubcommand := fmt.Sprintf("command %q", err.Command)
	if len(err.Path) > 0 {
		commandOrSubcommand = fmt.Sprintf("subcommand %q of %s", err.Command, strings.Join(err.Path, " "))
	}
	return fmt.Sprintf("%s has an empty string defined as an alias, which is invalid", commandOrSubcommand)
}

// CommandDuplicatesNameAsAliasError is returned when a [Command] has the same
// value used in its Name property in its Aliases property, which is invalid.
type CommandDuplicatesNameAsAliasError struct {
	// Path indicates the parents, if any, of the Command that duplicated
	// its Name property into its Aliases property. If Path is empty, that
	// indicates the Command is defined in Application.Commands. Otherwise,
	// the Command is defined in the Subcommands property of the Command
	// named by the last element of the slice.
	Path []string

	// Command is the name that was included as both a Name and in the
	// Aliases.
	Command string
}

func (err CommandDuplicatesNameAsAliasError) Error() string {
	commandOrSubcommand := fmt.Sprintf("command %q", err.Command)
	if len(err.Path) > 0 {
		commandOrSubcommand = fmt.Sprintf("subcommand %q of %q", err.Command, strings.Join(err.Path, " "))
	}
	return fmt.Sprintf("%s lists its own name as an alias, which is invalid", commandOrSubcommand)
}

// UnknownCommandError is returned when an invocation asks for a [Command] that
// has not been defined.
type UnknownCommandError struct {
	// Path is the list of commands that were invoked, with the Command
	// that has not been defined as the last element of the slice.
	Path []string
}

func (err UnknownCommandError) Error() string {
	return fmt.Sprintf("%q is not a valid command", strings.Join(err.Path, " "))
}

// InvalidCommandNameError is returned when a [Command] is defined with a Name
// or an Aliases key that is not a valid Command name.
type InvalidCommandNameError struct {
	// Name is the name or alias that's invalid.
	Name string

	// Path is list of parents, if any, that lead to the invalid Command,
	// including the invalid name.
	Path []string

	// Alias is set to true if the invalid name is defined in
	// Command.Aliases; if false, it's the Command.Name.
	Alias bool
}

func (err InvalidCommandNameError) Error() string {
	path := fmt.Sprintf("%q", err.Name)
	description := "defined at"
	if err.Alias {
		description = "alias of"
	}
	if len(err.Path) > 0 {
		path = fmt.Sprintf("%s (%s %s)", path, description, strings.Join(err.Path, " "))
	}
	return fmt.Sprintf("%s is not a valid command name, commands must contain only letters, numbers, -, _, and :, and cannot start with -", path)
}
