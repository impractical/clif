package clif

import (
	"context"
	"fmt"
	"maps"
	"strings"
)

// ExtraInputError is returned when a command gets more input than it knows
// what to do with.
type ExtraInputError struct {
	// Application is the Application that produced the error.
	Application Application
	// CommandPath is the Commands, in order, that were matched before the
	// error was produced. Each Command in the slice is the child of the
	// Command before it in the slice.
	CommandPath []Command
	// Flags holds the Flags that were matched before the error was
	// produced.
	Flags map[string]Flag
	// Args holds the positional arguments that were parsed before the
	// error was produced.
	Args []string
	// ExtraInput holds the extra, unexpected input.
	ExtraInput []string
}

func (err ExtraInputError) Error() string {
	var commandPath []string
	for _, cmd := range err.CommandPath {
		commandPath = append(commandPath, cmd.Name)
	}
	return fmt.Sprintf("unexpected extra input to %s: %s", strings.Join(commandPath, " "), strings.Join(err.ExtraInput, " "))
}

type parseable interface {
	subcommands() []Command
	flags() []FlagDef
	argsAccepted() bool
}

// RouteResult holds information about the [Command] that should be run and the
// Flags and arguments to pass to it, based on the parsing done by [Route].
type RouteResult struct {
	// Command is the Command that Route believes should be run.
	Command Command
	// Flags are the Flags that should be applied to that command.
	Flags map[string]Flag
	// Args are the positional arguments that should be passed to that
	// command.
	Args []string
}

// Route parses the passed input in the context of the passed [Application],
// turning it into a [Command] with Flags and arguments.
func Route(ctx context.Context, root Application, input []string) (RouteResult, error) {
	result := RouteResult{
		Flags: map[string]Flag{},
	}
	flagDefs := map[string]FlagDef{}
	flagList := listFlagDefs(root, true)
	for _, flag := range flagList {
		name := strings.ToLower(flag.Name)
		_, ok := flagDefs[name]
		if ok {
			return result, DuplicateFlagNameError(name)
		}
		flagDefs[name] = flag
		for _, alias := range flag.Aliases {
			alias = strings.ToLower(alias)
			_, ok := flagDefs[alias]
			if ok {
				return result, DuplicateFlagNameError(alias)
			}
			flagDefs[alias] = flag
		}
	}
	var cmdPath []Command
	parsed, err := parse(ctx, root, input, flagDefs, false)
	if err != nil {
		return result, err
	}
	maps.Copy(result.Flags, parsed.flags)
	result.Args = append(result.Args, parsed.args...)
	for parsed.subcommand != nil {
		result.Command = *parsed.subcommand
		cmdPath = append(cmdPath, *parsed.subcommand)
		parsed, err = parse(ctx, parsed.subcommand, parsed.unparsed, flagDefs, result.Command.AllowNonFlagFlags)
		if err != nil {
			return result, err
		}
		maps.Copy(result.Flags, parsed.flags)
		result.Args = append(result.Args, parsed.args...)
	}
	if len(parsed.unparsed) > 0 {
		return result, ExtraInputError{
			Application: root,
			CommandPath: cmdPath,
			Flags:       result.Flags,
			Args:        result.Args,
			ExtraInput:  parsed.unparsed,
		}
	}
	return result, nil
}
