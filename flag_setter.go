package clif

import (
	"context"
)

// FlagValueSetter is an interface to control how a [FlagValue] will be parsed
// into a type. The type can implement the SetFromFlagValue method to override
// how the [FlagSet.As], [FlagValues.As], and [FlagValue.As] methods will parse
// a flag value for that type.
type FlagValueSetter interface {
	// SetFromFlagValue updates the receiver to hold the FlagValue that was
	// passed in, returning an error if it cannot.
	SetFromFlagValue(ctx context.Context, value FlagValue) error
}

// FlagValuesSetter is an interface to control how a [FlagValues] will be
// parsed into a type. The type can implement the SetFromFlagValues method to
// override how the [FlagSet.As] and [FlagValues.As] methods will parse flag
// values for that type.
type FlagValuesSetter interface {
	// SetFromFlagValues updates the receiver to hold the FlagValues that
	// were passed in, returning an error if it cannot.
	SetFromFlagValues(ctx context.Context, values FlagValues) error
}

// FlagSetSetter is an interface to control how a [FlagSet] will be parsed into
// a type. The type can implement the SetFromFlagSet method to override how the
// [FlagSet.As] method will parse flag values for that type.
type FlagSetSetter interface {
	// SetFromFlagSet updates the receiver to hold the FlagSet that was
	// passed in, returning an error if it cannot.
	SetFromFlagSet(ctx context.Context, flags FlagSet) error
}
