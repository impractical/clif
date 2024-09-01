package clif

import (
	"strings"
	"text/tabwriter"
)

// SubcommandsHelp returns a default usage string for the subcommands provided
// by the passed [Command] or [Application].
func SubcommandsHelp(command parseable) string {
	var builder strings.Builder
	writer := tabwriter.NewWriter(&builder, 4, 4, 1, '\t', 0) //nolint:mnd // 4 spaces to a tab is just magic, dunno what to say
	for _, cmd := range command.subcommands() {
		writer.Write([]byte(cmd.Name + "\t" + cmd.Description + "\n")) //nolint:errcheck // error shouldn't be possible here
	}
	writer.Flush() //nolint:errcheck // error shouldn't be possible here
	return builder.String()
}

// FlagsHelp returns a default usage string for the flags defined for the
// passed [Command] or [Application].
func FlagsHelp(command parseable) string {
	var builder strings.Builder
	writer := tabwriter.NewWriter(&builder, 4, 4, 1, '\t', 0) //nolint:mnd // 4 spaces to a tab is just magic, dunno what to say
	for _, flag := range command.flags() {
		writer.Write([]byte(flag.Name + "\t<" + flag.Parser.FlagType() + ">\t" + flag.Description + "\n")) //nolint:errcheck // error shouldn't be possible here
	}
	writer.Flush() //nolint:errcheck // error shouldn't be possible here
	return builder.String()
}
