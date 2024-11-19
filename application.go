package clif

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	// ErrAppHasNoCommands is returned when an application does not define
	// any commands, which is invalid.
	ErrAppHasNoCommands = errors.New("application does not define any commands")
)

// Application is the root definition of a CLI.
type Application struct {
	// Commands are the commands that the application supports.
	Commands []Command

	// Flags are the definitions for any global flags the application
	// supports.
	Flags []FlagDef

	Handler HandlerBuilder
}

// Run executes the invoked command. It routes the input to the appropriate
// [Command], parses it with the [HandlerBuilder], and executes the [Handler].
// The return is the status code the command has indicated it exited with.
func (app Application) Run(ctx context.Context, opts ...RunOption) int {
	options := RunOptions{
		Output: os.Stdout,
		Error:  os.Stderr,
		Args:   os.Args[1:],
	}
	for _, opt := range opts {
		opt(&options)
	}
	resp := &Response{
		Output: options.Output,
		Error:  options.Error,
		Code:   0,
	}
	// Route parses out the distinct parts of our input and finds the right
	// command to execute them.
	result, err := Route(ctx, app, options.Args)
	if err != nil {
		fmt.Fprintln(resp.Error, err.Error()) //nolint:errcheck // if there's an error, we can't do anything
		return 1
	}

	if result.Command.Handler == nil {
		fmt.Fprintln(resp.Error, "invalid command:", strings.Join(options.Args, " ")) //nolint:errcheck // if there's an error, we can't do anything
		return 1
	}

	// Build makes us a handler, parsing all the input and injecting it
	// into a handler-specific format
	handler := result.Command.Handler.Build(ctx, result.Flags, result.Args, resp)
	if resp.Code > 0 {
		return resp.Code
	}

	// Handle executes the handler
	handler.Handle(ctx, resp)
	return resp.Code
}

// Validate determines whether an [Application] has a valid definition or not.
func (app Application) Validate(ctx context.Context) error {
	var errs error
	if len(app.Commands) < 1 {
		errs = errors.Join(errs, ErrAppHasNoCommands)
	}
	flagKeys := map[string]struct{}{}
	for _, flagDef := range app.Flags {
		if _, ok := flagKeys[flagDef.Name]; ok {
			errs = errors.Join(errs, DuplicateFlagNameError(flagDef.Name))
		}
		flagKeys[flagDef.Name] = struct{}{}
		for _, alias := range flagDef.Aliases {
			if _, ok := flagKeys[alias]; ok {
				errs = errors.Join(errs, DuplicateFlagNameError(alias))
			}
		}
	}
	cmdNames := map[string]struct{}{}
	for pos, cmd := range app.Commands {
		if cmd.Name == "" {
			errs = errors.Join(errs, CommandMissingNameError{Path: []string{}, Pos: pos})
			continue
		}
		if _, ok := cmdNames[cmd.Name]; ok {
			errs = errors.Join(errs, DuplicateCommandError{Path: []string{}, Command: cmd.Name})
		}
		for _, alias := range cmd.Aliases {
			if alias == "" {
				errs = errors.Join(errs, CommandAliasEmptyError{Path: []string{}, Command: cmd.Name})
			} else if alias == cmd.Name {
				errs = errors.Join(errs, CommandDuplicatesNameAsAliasError{Path: []string{}, Command: cmd.Name})
			} else if _, ok := cmdNames[alias]; ok {
				errs = errors.Join(errs, DuplicateCommandError{Path: []string{}, Command: alias})
			}
		}
		errs = errors.Join(errs, cmd.Validate(ctx, []string{cmd.Name}, flagKeys))
	}
	return errs
}
