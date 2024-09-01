package clif

import (
	"context"
	"fmt"
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
	// this command, separate from flag values and subcommands.
	ArgsAccepted bool

	// AllowNonFlagFlags controls whether things that aren't flags (like
	// flag values, subcommands, and arguments) can start with --. If
	// false, we'll throw an error when we encounter an -- that doesn't
	// have a FlagDef for it on this command or any of its subcommands. If
	// true, we'll allow it, using it either as a flag value, subcommand,
	// or argument, whichever is allowed. If none are allowed, it will
	// still throw an invalid flag error.
	AllowNonFlagFlags bool
}

func (cmd Command) argsAccepted() bool     { return cmd.ArgsAccepted }
func (cmd Command) subcommands() []Command { return cmd.Subcommands }
func (cmd Command) flags() []FlagDef       { return cmd.Flags }

type parsedCommand struct {
	subcommand *Command
	flags      map[string]Flag
	args       []string
	unparsed   []string
}

func parse(ctx context.Context, root parseable, args []string, allowNonFlagFlags bool) (parsedCommand, error) {
	res := parsedCommand{
		flags: map[string]Flag{},
	}
	if len(args) < 1 {
		return res, nil
	}
	allFlags := map[string]FlagDef{}
	flagList := listFlagDefs(root, true)
	for _, flag := range flagList {
		name := strings.ToLower(flag.Name)
		_, ok := allFlags[name]
		if ok {
			return res, DuplicateFlagNameError(name)
		}
		allFlags[name] = flag
		for _, alias := range flag.Aliases {
			alias = strings.ToLower(alias)
			_, ok := allFlags[alias]
			if ok {
				return res, DuplicateFlagNameError(alias)
			}
			allFlags[alias] = flag
		}
	}
	var openFlagDef *FlagDef
	var openFlagArg string
	for pos, arg := range args {
		// if this argument matches a flag definition we're expecting,
		// let's assume it's that flag definition. In theory it could
		// be the argument to the open flag definition and just
		// coincidentally match, or it could be an argument to the
		// command or one of its subcommands, but it's probably fair to
		// ask consumers to not allow that confusion to exist.
		if strings.HasPrefix(arg, "--") {
			trimmed := strings.TrimPrefix(arg, "--")
			argument, value, hasValue := strings.Cut(trimmed, "=")
			arg = strings.ToLower(argument)
			flagDef, ok := allFlags[arg]
			if ok {
				// if we've declared another flag but there's an open
				// flag definition, it has no value, close it
				if openFlagDef != nil {
					flag, err := openFlagDef.Parser.Parse(ctx, openFlagArg, "", res.flags[openFlagArg])
					if err != nil {
						return res, err
					}
					res.flags[flag.GetName()] = flag
					openFlagDef = nil
					openFlagArg = ""
				}

				// if the flag definition doesn't accept values
				// but we have a key=value argument for that
				// flag, this isn't a valid invocation
				if !flagDef.ValueAccepted && hasValue {
					return res, UnexpectedFlagValueError{Flag: arg, Value: value}
				}

				// if this flag doesn't accept values, or we
				// already have the value, parse it and we're
				// done with this argument
				if !flagDef.ValueAccepted || hasValue {
					// TODO: for flags that can be specified multiple times, we need to pass in the existing value so it can be modified
					flag, err := flagDef.Parser.Parse(ctx, arg, value, res.flags[arg])
					if err != nil {
						return res, err
					}
					res.flags[flag.GetName()] = flag
					continue
				}

				// if this flag doesn't have a value yet, it's
				// an open flag value. Move on to the next arg,
				// which may be this flag's value.
				if !hasValue {
					// we have a flag that accepts a value but
					// there isn't one in this arg. The next arg
					// must be the value
					openFlagDef = &flagDef
					openFlagArg = arg
					continue
				}
			} else if !allowNonFlagFlags {
				// if it doesn't match one of our flag definitions and
				// we don't allow that, it's an error
				return res, UnknownFlagNameError(arg)
			}
		}

		lowerArg := strings.ToLower(arg)

		// this is now either the optional value to the open flag
		// definition (if there is one), a subcommand, or an argument
		// to the command.

		// let's eliminate subcommand as a possibility, because that's
		// a pretty closed set.
		for _, sub := range root.subcommands() {
			var match bool
			if lowerArg == strings.ToLower(sub.Name) {
				match = true
			} else {
				for _, alias := range sub.Aliases {
					if lowerArg == strings.ToLower(alias) {
						match = true
						break
					}
				}
			}
			if match {
				// if there's still an open flag definition, it
				// has no value, we have a subcommand instead.
				//
				// in theory, if a flag's value was the same as
				// a valid subcommand, this would confuse the
				// flag's value for the subcommand. But it
				// seems reasonable to expect consumers to not
				// allow that confusion.
				if openFlagDef != nil {
					flag, err := openFlagDef.Parser.Parse(ctx, openFlagArg, "", res.flags[openFlagArg])
					if err != nil {
						return res, err
					}
					res.flags[flag.GetName()] = flag
				}
				res.subcommand = &sub
				if len(args) > pos+1 {
					res.unparsed = args[pos+1:]
				}
				return res, nil
			}
		}

		// this is either an optional value to the open flag definition
		// (if there is one) or an argument to the command. If we don't
		// have an open flag definition and don't accept args, this
		// isn't a valid invocation.
		if !root.argsAccepted() && openFlagDef == nil {
			return res, UnexpectedCommandArgError(arg)
		}

		// if we don't accept args and have an open flag definition,
		// assume this is the flag's value.
		if !root.argsAccepted() {
			flag, err := openFlagDef.Parser.Parse(ctx, openFlagArg, arg, res.flags[openFlagArg])
			if err != nil {
				return res, err
			}
			res.flags[flag.GetName()] = flag
			openFlagDef = nil
			openFlagArg = ""
			continue
		}

		// if we don't have an open flag definition, assume this is an
		// argument to the command
		if openFlagDef == nil {
			res.args = append(res.args, arg)
			continue
		}

		// we have an open flag definition and we accept arguments.
		// This could be either. Let's assume, if this is the last
		// argument, that it's a command argument. Otherwise, we're
		// assuming it's a flag value.
		if pos == len(args)-1 {
			res.args = append(res.args, arg)
			continue
		}

		flag, err := openFlagDef.Parser.Parse(ctx, openFlagArg, arg, res.flags[openFlagArg])
		if err != nil {
			return res, err
		}
		res.flags[flag.GetName()] = flag
		openFlagDef = nil
		openFlagArg = ""
		continue
	}
	return res, nil
}
