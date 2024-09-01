package flagtypes

// ListFlag implements [clif.Flag] as a flag that can be specified multiple
// times.
type ListFlag[FlagType BasicFlagConstraint] struct {
	// Name will be set to the name the flag was invoked with.
	Name string

	// RawValue will be set to the string the user passed.
	RawValue string

	// Value will be set to the value that RawValue parsed into.
	Value []FlagType
}

// GetName fills the [clif.Flag] interface and returns the name the flag was
// invoked with.
func (flag ListFlag[FlagType]) GetName() string {
	return flag.Name
}

// GetRawValue fills the [clif.Flag] interface and returns the string the user
// passed as the flag's value.
func (flag ListFlag[FlagType]) GetRawValue() string {
	return flag.RawValue
}
