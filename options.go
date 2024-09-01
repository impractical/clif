package clif

import (
	"io"
)

// RunOptions holds all the options to pass to [Application.Run]. It should
// be built by using [RunOption]s to modify a passed in RunOptions.
type RunOptions struct {
	// Output is where command output should be written. Defaults to
	// [os.Stdout].
	Output io.Writer

	// Error is where the command should write information about errors.
	// Defaults to [os.Stderr].
	Error io.Writer

	// Args are the arguments that were passed to the command. Defaults
	// to [os.Args][1:].
	Args []string
}

// RunOption is a function type that modifies a passed [RunOptions] when
// called. It's used to configure the behavior of [Application.Run].
type RunOption func(*RunOptions)

// WithOutput is a [RunOption] that sets the command output to the passed
// [io.Writer].
func WithOutput(w io.Writer) RunOption {
	return func(opts *RunOptions) {
		opts.Output = w
	}
}

// WithError is a [RunOption] that sets the [io.Writer] the application will
// write information about errors to to the passed [io.Writer].
func WithError(w io.Writer) RunOption {
	return func(opts *RunOptions) {
		opts.Error = w
	}
}

// WithArgs is a [RunOption] that sets the arguments that will be parsed as the
// command's input to the passed strings.
func WithArgs(args []string) RunOption {
	return func(opts *RunOptions) {
		opts.Args = args
	}
}
