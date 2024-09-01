package flagtypes

import (
	"time"
)

// BasicFlagConstraint describes the types that the [BasicFlag] [Flag]
// implementation supports.
type BasicFlagConstraint interface {
	~bool |
		~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 |
		~complex64 | ~complex128 |
		time.Time
}

// BasicFlag implements [Flag] for a base set of builtin types, allowing out of the box functionality similar to the [flag] package.
type BasicFlag[FlagType BasicFlagConstraint] struct {
	// Name will be set to the name the flag was invoked with.
	Name string

	// RawValue will be set to the string the user passed.
	RawValue string

	// Value will be set to the value that RawValue parsed into.
	Value FlagType
}

// GetName fills the [Flag] interface and returns the name the flag was invoked
// with.
func (flag BasicFlag[FlagType]) GetName() string {
	return flag.Name
}

// GetRawValue fills the [Flag] interface and returns the string the user
// passed as the flag's value.
func (flag BasicFlag[FlagType]) GetRawValue() string {
	return flag.RawValue
}
