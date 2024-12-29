package clif

import (
	"context"
	"os"
	"strings"
)

// RouteResult holds information about the [Command] that should be run and the
// [FlagSet] and arguments to pass to it, based on the parsing done by
// [Route].
type RouteResult struct {
	// Command is the Command that Route believes should be run. If nil, no
	// command was passed and the top-level Application handler should be
	// called. A nil value does not indicate a command was specified but
	// couldn't be found; that will be an explicit error.
	Command *Command

	// CommandPath is the list of Commands that were invoked to reach
	// Command, including Command itself.
	CommandPath []Command

	// Flags are the flags that should be applied to that command.
	Flags FlagSet

	// DefinedFlags are the FlagDefs that are defined along the
	// CommandPath.
	DefinedFlags []FlagDef

	// Args are the positional arguments that should be passed to that
	// command.
	Args []string
}

// Route parses the passed input in the context of the passed [Application],
// turning it into a [Command] with a [FlagSet] and arguments.
func Route(ctx context.Context, app Application, input []string) (RouteResult, error) {
	err := app.Validate(ctx)
	if err != nil {
		return RouteResult{}, err
	}
	parser := newParser(ctx, input)
	parser.mark(ctx)
	parser.normalize(ctx)
	commandDefs := inputParserCommand{
		subcommands: map[string]inputParserCommand{},
		flags:       map[string]inputParserFlag{},
	}
	for _, flagDef := range app.Flags {
		commandDefs.flags[flagDef.Name] = inputParserFlag{
			allowsValue:   !flagDef.OnlyToggle,
			mustHaveValue: !flagDef.IsToggle,
		}
		for _, alias := range flagDef.Aliases {
			commandDefs.flags[alias] = inputParserFlag{
				allowsValue:   !flagDef.OnlyToggle,
				mustHaveValue: !flagDef.IsToggle,
			}
		}
	}
	for _, cmd := range app.Commands {
		cmdDef := getCommandDef(cmd)
		commandDefs.subcommands[cmd.Name] = cmdDef
		for _, alias := range cmd.Aliases {
			commandDefs.subcommands[alias] = cmdDef
		}
	}
	err = parser.apply(ctx, commandDefs)
	if err != nil {
		return RouteResult{}, err
	}
	router := inputRouter{
		// a stub command to start our command tree off, just the
		// top-level subcommands of the CLI itself
		cmd: &Command{
			// give this command an invalid name so we can detect
			// if it's still the active command after router.route
			// runs
			Name:        "--placeholder--",
			Subcommands: app.Commands,
		},
		// gotta initialize the flags map so we don't panic on write
		flags:  map[string][]*string{},
		parser: parser,
	}

	err = router.route(ctx)
	if err != nil {
		return RouteResult{}, err
	}

	if router.cmd != nil && router.cmd.Name == "--placeholder--" {
		router.cmd = nil
	}

	// build a map of flags the commands we visited actually accept so we
	// can validate the passed flags are acceptable
	acceptedFlags := map[string]FlagDef{}
	var cmdPath []Command
	var resultFlagDefs []FlagDef
	for _, def := range app.Flags {
		acceptedFlags[def.Name] = def
		resultFlagDefs = append(resultFlagDefs, def)
		for _, alias := range def.Aliases {
			acceptedFlags[alias] = def
		}
	}
	for _, entry := range router.path {
		cmdPath = append(cmdPath, entry.cmd)
		for _, def := range entry.cmd.Flags {
			acceptedFlags[def.Name] = def
			resultFlagDefs = append(resultFlagDefs, def)
			for _, alias := range def.Aliases {
				acceptedFlags[alias] = def
			}
		}
	}

	// pull out the flag value into a Set we can return, and check
	// that they're valid as we do
	flagValues := FlagSet{}
	for flag, values := range router.flags {
		def, ok := acceptedFlags[flag]
		if !ok {
			// if we don't have a definition for this flag, that's
			// a problem
			return RouteResult{}, UnknownFlagNameError(flag)
		}
		flagWithoutLeadingHyphens := strings.TrimPrefix(strings.TrimPrefix(flag, "-"), "-")
		nameWithoutLeadingHyphens := strings.TrimPrefix(strings.TrimPrefix(def.Name, "-"), "-")
		for _, value := range values {
			val := FlagValue{
				HasValue:     value != nil,
				Key:          flagWithoutLeadingHyphens,
				CanonicalKey: nameWithoutLeadingHyphens,
			}
			if value != nil {
				// if we have a non-nil value but this flag can
				// only be used as a toggle, that's a problem
				if def.OnlyToggle {
					return RouteResult{}, UnexpectedFlagValueError{
						Flag:  flag,
						Value: *value,
					}
				}
				val.Raw = *value
			} else if !def.IsToggle {
				// if we don't have a value but this flag can't
				// be used as a toggle, that's a problem
				return RouteResult{}, MissingFlagValueError(flag)
			}
			flagValues[nameWithoutLeadingHyphens] = append(flagValues[nameWithoutLeadingHyphens], val)
		}
		if len(flagValues[nameWithoutLeadingHyphens]) > 1 && !def.AllowMultiple {
			// if we have more than one value but only accept a
			// single value, that's a problem
			return RouteResult{}, TooManyFlagValuesError(flag)
		}
	}

	// we know our specified flag values were valid, let's fill in the
	// unspecified flag values with defaults now
	for key, def := range acceptedFlags {
		if key != def.Name {
			// don't deal with aliases, we only want to process
			// this once per flag
			continue
		}

		keyWithoutLeadingHyphens := strings.TrimPrefix(strings.TrimPrefix(key, "-"), "-")

		if len(flagValues[keyWithoutLeadingHyphens]) > 0 {
			// if we already have a value, there's no need to fall
			// back, let's just move on
			continue
		}

		// if we have defined environment variables to fall back on,
		// now's the time to fall back on those environment variables
		for _, envVar := range def.FromEnvVars {
			envVal := os.Getenv(envVar)
			if envVal != "" {
				flagValues[keyWithoutLeadingHyphens] = append(flagValues[keyWithoutLeadingHyphens], FlagValue{
					HasValue:     true,
					Raw:          envVal,
					CanonicalKey: strings.TrimPrefix(strings.TrimPrefix(def.Name, "-"), "-"),
				})
				break
			}
		}

		if len(flagValues[keyWithoutLeadingHyphens]) > 0 {
			// if the env var set a value, we're done here
			continue
		}

		// if we've defined Default values, go ahead and use them
		if len(def.Default) > 0 {
			flagValues[keyWithoutLeadingHyphens] = def.Default
		}
	}

	return RouteResult{
		Command:      router.cmd,
		CommandPath:  cmdPath,
		Flags:        flagValues,
		DefinedFlags: resultFlagDefs,
		Args:         router.args,
	}, nil
}

func getCommandDef(cmd Command) inputParserCommand {
	def := inputParserCommand{
		subcommands: map[string]inputParserCommand{},
		allowsArgs:  cmd.ArgsAccepted,
		flags:       map[string]inputParserFlag{},
	}
	for _, flagDef := range cmd.Flags {
		def.flags[flagDef.Name] = inputParserFlag{
			allowsValue:   !flagDef.OnlyToggle,
			mustHaveValue: !flagDef.IsToggle,
		}
		for _, alias := range flagDef.Aliases {
			def.flags[alias] = inputParserFlag{
				allowsValue:   !flagDef.OnlyToggle,
				mustHaveValue: !flagDef.IsToggle,
			}
		}
	}
	for _, sub := range cmd.Subcommands {
		subDef := getCommandDef(sub)
		def.subcommands[sub.Name] = subDef
		for _, alias := range sub.Aliases {
			def.subcommands[alias] = subDef
		}
	}
	return def
}
