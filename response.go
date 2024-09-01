package clif

import (
	"io"
)

// Response holds the ways a command can present information to the user.
type Response struct {
	// Code is the status code the command will exit with.
	Code int

	// Output is the writer that should be used for command output. It will
	// usually be set to the shell's standard output.
	Output io.Writer

	// Error is the writer that should be used to communicate error
	// conditions. It will usually be set to the shell's standard error.
	Error io.Writer
}
