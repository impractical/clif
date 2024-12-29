package clif

import (
	"context"
)

// Middleware provides an interface for specifying code to run after the
// appropriate [Handler] to execute is identified but before the [Handler] is
// executed.
type Middleware interface {
	// Intercept is the code that will run when the Middleware is executed.
	// All the runtime parameters available to a Handler are available,
	// along with DefParams describing the Application, the Command to
	// execute (if any), the parents of that Command (if any), and the
	// acceptable flags for that particular Command.
	//
	// If Intercept returns false, execution will halt and the Handler and
	// any other Middleware will not be executed.
	Intercept(ctx context.Context, flags FlagSet, args []string, defs DefParams, resp *Response) bool
}
