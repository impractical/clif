package clif

// DefParams holds a parameter bundle of the definitions that are relevant to a
// given execution. A definition is relevant to a given execution if it's a
// [Command] in the command path of a [Command] that has been routed to, the
// [Command] that has been routed to, the root [Application], or a [FlagDef]
// that is set on any of those [Command]s or the [Application].
type DefParams struct {
	CommandPath     []Command
	Command         *Command
	AcceptableFlags []FlagDef
	App             Application
}
