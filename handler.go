package clif

import (
	"context"
)

// Handler is an interface that commands should implement. The implementing
// type should probably be a struct, with arguments and dependencies defined as
// fields on the struct.
type Handler interface {
	// Handle is a method that will be called when the command is executed.
	// It should contain the business logic of the command.
	Handle(ctx context.Context, resp *Response)
}

// HandlerBuilder is an interface that should wrap a [Handler]. It parses the
// passed [Flag]s and args into a [Handler], to separate out the parsing logic
// from the business logic.
type HandlerBuilder interface {
	// Build creates a Handler by parsing the Flags and args into the
	// appropriate handler type.
	Build(ctx context.Context, flags []Flag, args []string, resp *Response) Handler
}

// TODO: should flags be a map, not a slice?
