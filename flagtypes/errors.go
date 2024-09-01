package flagtypes

import (
	"fmt"
)

// UnexpectedFlagPriorTypeError is returned when a FlagParser is passed a prior
// value for that flag that wasn't of the type it was expecting. This shouldn't
// be possible unless the FlagParser returns Flags it doesn't expect to receive
// as prior values.
type UnexpectedFlagPriorTypeError struct {
	Name     string
	Expected any
	Got      any
}

func (err UnexpectedFlagPriorTypeError) Error() string {
	return fmt.Sprintf("expected prior value of flag %q to be %T, got %T", err.Name, err.Expected, err.Got)
}

// UnexpectedFlagValueTypeError is returned when a [ListFlag] parser is relying
// on another [FlagParser] to parse each flag in the collection, but doesn't
// get the [Flag] type it expects. This usually indicates a bug in the
// [ListFlag] parser.
type UnexpectedFlagValueTypeError struct {
	Name     string
	Expected any
	Got      any
}

func (err UnexpectedFlagValueTypeError) Error() string {
	return fmt.Sprintf("expected value of flag %q to be %T, got %T", err.Name, err.Expected, err.Got)
}
