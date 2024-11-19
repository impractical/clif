package clif

import (
	"fmt"
	"reflect"
)

// InvalidKindError is returned when a flag value cannot be parsed into a type
// because the kind of type it is isn't supported.
type InvalidKindError struct {
	// Value holds the value that the caller was attempting to parse into.
	Value any

	// Kind indicates the reflect.Kind of the value that the caller was
	// attempting to parse into.
	Kind reflect.Kind
}

func (err InvalidKindError) Error() string {
	return fmt.Sprintf("cannot parse into %T (kind: %s)", err.Value, err.Kind)
}

// InvalidConversionError is returned when a [FlagValue] is parsed into a type
// that isn't supported.
type InvalidConversionError struct {
	// Source is the FlagValue that was parsed.
	Source FlagValue

	// Target is the value being parsed into.
	Target reflect.Value
}

func (err InvalidConversionError) Error() string {
	return fmt.Sprintf("don't know how to parse flag value %q into %s", err.Source.Raw, err.Target.Type())
}

// FlagKeyNotInStructError is returned when a flag key is passed during
// invocation that isn't found in the struct being parsed into. This usually
// indicates that a struct doesn't have a field for every defined flag; flags
// with no [FlagDef] will return an [UnknownFlagNameError] instead.
type FlagKeyNotInStructError struct {
	// Key is the flag key that was not defined in the struct.
	Key string

	// StructType is the package-delimited type of the struct that was
	// missing the flag key.
	StructType string
}

func (err FlagKeyNotInStructError) Error() string {
	return fmt.Sprintf("flag key %q cannot be parsed into %s, has no field can hold it", err.Key, err.StructType)
}

// FlagSetterResponseError is returned when the SetFromFlagSet,
// SetFromFlagValues, or SetFromFlagValue method is called on a
// [FlagValueSetter], [FlagValuesSetter], or FlagSetSetter], and the return
// value is somehow does not fill the error interface.
//
// This pretty much always means there's a bug in clif.
type FlagSetterResponseError struct {
	// Receiver is the value the method was called on.
	Receiver any

	// Method is the name of the method that was called.
	Method string

	// Value is the value that was returned.
	Value any
}

func (err FlagSetterResponseError) Error() string {
	return fmt.Sprintf("the %s method on %T did not return an error as expected, it returned a %T with value %v -- this usually indicates an error in clif", err.Method, err.Receiver, err.Value, err.Value)
}

// FlagSetterMethodInvalidError is returned when the SetFromFlagSet method
// can't be found on a [FlagSetSetter], the SetFromFlagValues method can't be
// found on a [FlagValuesSetter], or the SetFromFlagValue method can't be found
// on a [FlagValueSetter]. This shouldn't ever happen, and if it does, it
// probably means clif's reflection code has drifted from the interface type
// definitions.
//
// If you encounter this, report it as a bug.
type FlagSetterMethodInvalidError struct {
	// Receiver is the value the method was called on.
	Receiver any

	// Method is the name of the method that was called.
	Method string
}

func (err FlagSetterMethodInvalidError) Error() string {
	return fmt.Sprintf("the %s method on %T wasn't found, this almost always indicates a bug in clif", err.Method, err.Receiver)
}

// StructFieldMissingFlagTagError is returned when parsing a [FlagSet] into a
// struct type using [FlagSet.As], and one of the struct's fields doesn't have
// a "flag" struct tag on it. All exported fields of the struct must have a
// "flag" struct tag.
type StructFieldMissingFlagTagError struct {
	// Struct is a value of the struct type that the field belongs to.
	Struct any

	// Field is the name of the field missing the tag.
	Field string
}

func (err StructFieldMissingFlagTagError) Error() string {
	return fmt.Sprintf("the %s field on %T needs a struct tag for %q -- if you don't want to parse flags into it, set it to %q", err.Field, err.Struct, "flag", "-")
}
