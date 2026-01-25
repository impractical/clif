package clif

import (
	"context"
	"strings"

	"github.com/mitchellh/go-wordwrap"
)

// HelpMiddleware provides a middleware that, when any of the specified flags
// are passed, will print the help output and prevent any other [Handler] or
// [Middleware] from executing.
type HelpMiddleware struct {
	Flags []string
}

func (middleware HelpMiddleware) Intercept(ctx context.Context, flags FlagSet, _ []string, defs DefParams, resp *Response) bool { //nolint:revive // that's the way interfaces go sometimes, this really needs 5 arguments
	var showHelp bool
	for _, flag := range middleware.Flags {
		helpFlag, ok := flags[flag]
		if !ok {
			continue
		}
		err := helpFlag.As(ctx, &showHelp)
		if err != nil {
			fprintlnOrPanic(resp.Error, "Error checking if the", helpFlag, "flag is set:", err.Error())
			resp.Code = 1
			return false
		}
		if showHelp {
			break
		}
	}
	if !showHelp {
		return true
	}
	HelpHandler{
		App:     defs.App,
		CmdPath: defs.CommandPath,
		Cmd:     defs.Command,
		Flags:   defs.AcceptableFlags,
	}.Handle(ctx, resp)
	return false
}

// HelpHandler is a [Handler] that writes help output to the [Response.Output].
// The help output is not guaranteed to be stable between versions of clif.
type HelpHandler struct {
	App     Application
	CmdPath []Command
	Cmd     *Command
	Flags   []FlagDef
}

// Handle is a method that will be called when the command is executed.
// It should contain the business logic of the command.
func (handler HelpHandler) Handle(ctx context.Context, resp *Response) {
	if handler.Cmd == nil {
		handler.applicationHelp(ctx, resp)
		return
	}
	handler.commandHelp(ctx, resp)
}

func (handler HelpHandler) applicationHelp(_ context.Context, resp *Response) {
	fprintOrPanic(resp.Output, handler.App.Name)
	if handler.App.Version != "" {
		fprintOrPanic(resp.Output, " ", handler.App.Version)
	}
	if len(handler.App.Description) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, wrapAndPad(handler.App.Description))
	}
	if len(handler.App.DetailedDescription) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "DESCRIPTION")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, wrapAndPad(handler.App.DetailedDescription))
	}
	var commands []Command
	for _, cmd := range handler.App.Commands {
		if cmd.Hidden {
			continue
		}
		commands = append(commands, cmd)
	}
	if len(commands) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "COMMANDS")
		fprintOrPanic(resp.Output, "\n")
		var commandTable [][2]string
		for _, cmd := range commands {
			commandTable = append(commandTable, [2]string{cmd.Name, cmd.Description})
		}
		fprintOrPanic(resp.Output, makeWrappedAndPaddedTable(commandTable))
	}
	if len(handler.Flags) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "FLAGS")
		fprintOrPanic(resp.Output, "\n")
		var flagTable [][2]string
		for _, flag := range handler.Flags {
			flagTable = append(flagTable, [2]string{flag.Name, flag.Description})
		}
		fprintOrPanic(resp.Output, makeWrappedAndPaddedTable(flagTable))
	}
	fprintOrPanic(resp.Output, "\n")
}

func (handler HelpHandler) commandHelp(_ context.Context, resp *Response) {
	if handler.App.Name != "" {
		fprintOrPanic(resp.Output, handler.App.Name, " ")
	}
	for pos, cmd := range handler.CmdPath {
		fprintOrPanic(resp.Output, cmd.Name)
		if len(handler.CmdPath) > pos+1 {
			fprintOrPanic(resp.Output, " ")
		}
	}
	if len(handler.Cmd.Description) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, wrapAndPad(handler.Cmd.Description))
	}
	if len(handler.Cmd.UsageExample) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "USAGE")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, wrapAndPad(handler.Cmd.UsageExample))
	}
	if len(handler.Cmd.DetailedDescription) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "DESCRIPTION")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, wrapAndPad(handler.Cmd.DetailedDescription))
	}
	var commands []Command
	for _, cmd := range handler.Cmd.Subcommands {
		if cmd.Hidden {
			continue
		}
		commands = append(commands, cmd)
	}
	if len(commands) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "SUBCOMMANDS")
		fprintOrPanic(resp.Output, "\n")
		var commandTable [][2]string
		for _, cmd := range commands {
			commandTable = append(commandTable, [2]string{cmd.Name, cmd.Description})
		}
		fprintOrPanic(resp.Output, makeWrappedAndPaddedTable(commandTable))
	}
	if len(handler.Flags) > 0 {
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "\n")
		fprintOrPanic(resp.Output, "FLAGS")
		fprintOrPanic(resp.Output, "\n")
		var flagTable [][2]string
		for _, flag := range handler.Flags {
			flagTable = append(flagTable, [2]string{flag.Name, flag.Description})
		}
		fprintOrPanic(resp.Output, makeWrappedAndPaddedTable(flagTable))
	}
	fprintOrPanic(resp.Output, "\n")
}

func wrapAndPad(input string) string {
	pad := "  "
	limit := uint(80) //nolint:mnd // 80 is a legitimate magic number, because terminals are weird
	if uint(len(pad)) > limit {
		panic("padding exceeds limit")
	}
	limit = limit - uint(len(pad))
	wrapped := wordwrap.WrapString(input, limit)
	padded := strings.ReplaceAll(wrapped, "\n", "\n"+pad)
	// strip the trailing pad
	padded = strings.TrimSuffix(padded, pad)
	// make sure the first line gets padded
	padded = pad + padded
	return padded
}

func makeWrappedAndPaddedTable(input [][2]string) string {
	pad := "  "
	limit := uint(80) //nolint:mnd // 80 is a legitimate magic number, because terminals are weird
	if uint(len(pad)) > limit {
		panic("padding exceeds limit")
	}
	limit = limit - uint(len(pad))
	var maxCol1 int
	for _, row := range input {
		if len(row[0]) > maxCol1 {
			maxCol1 = len(row[0])
		}
	}
	if uint(maxCol1+len(pad)+len(pad)) >= limit {
		panic("padding and first column exceeds limit")
	}
	limit = limit - uint(maxCol1+len(pad)+len(pad))
	var result strings.Builder
	for _, row := range input {
		result.WriteString(pad)
		result.WriteString(row[0])
		for range maxCol1 - len(row[0]) {
			result.WriteString(" ")
		}
		wrapped := wordwrap.WrapString(row[1], limit)
		wrappedLines := strings.Split(wrapped, "\n")
		for pos, line := range wrappedLines {
			if pos != 0 {
				result.WriteString(pad)
				for range maxCol1 {
					result.WriteString(" ")
				}
			}
			result.WriteString("    ")
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	return strings.TrimSuffix(result.String(), "\n")
}
