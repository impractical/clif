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
// passed [Flags] and args into a [Handler], to separate out the parsing logic
// from the business logic.
type HandlerBuilder interface {
	// Build creates a Handler by parsing the flags and args into the
	// appropriate handler type.
	Build(ctx context.Context, flags FlagSet, args []string, resp *Response) Handler
}

// DefInjector is an interface that a [HandlerBuilder] can optionally
// implement. If implemented, the InjectDefs method will be called to make
// the [DefParams] available at runtime to the [HandlerBuilder] before
// [HandlerBuilder.Build] is called.
type DefInjector interface {
	HandlerBuilder
	InjectDefs(ctx context.Context, defs DefParams, resp *Response)
}
