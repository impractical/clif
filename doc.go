// Package clif provides a framework for writing command line applications.
//
// The framework starts with an [Application], which defines some global flags
// that apply to all commands and the different Commands that the application
// accepts.
//
// Each flag is defined by a [FlagDef], which describes the flag name and how
// to parse it.
//
// Each [Command] describes the command name, any subcommands and flags it
// accepts, and other information about parsing the command and how to execute
// it.
//
// Once input is matched to the [Command], it calls the [HandlerBuilder]
// associated with that [Command]. The [HandlerBuilder] is responsible for
// turning flags, arguments, and a [Command] into a [Handler]. It's separated
// out from the [Handler] so the business logic of the [Handler] can be
// separated out from the logic to parse the flags and arguments.
//
// Finally, once we have a [Handler], it gets executed, with a [Response] to
// write output to and record the desired exit code of the command.
package clif
