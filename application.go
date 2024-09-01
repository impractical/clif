package clif

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Application is the root definition of a CLI.
type Application struct {
	// Commands are the commands that the application supports.
	Commands []Command

	// Flags are the definitions for any global flags the application
	// supports.
	Flags []FlagDef
}

func (Application) argsAccepted() bool         { return false }
func (app Application) subcommands() []Command { return app.Commands }
func (app Application) flags() []FlagDef       { return app.Flags }

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
