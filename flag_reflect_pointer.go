package clif

import (
	"context"
	"reflect"
)

// recursivelyPopulatePointer unwraps any number of pointers to get to the
// root value being pointed at. It then creates a new zero value that matches
// that type, and re-wraps it in the same number of pointers.
//
// The end result of this is a reflect.Value that can be set and used without
// worrying about nil pointer panics.
func recursivelyPopulatePointer(_ context.Context, target reflect.Value) reflect.Value {
	pointer := target.Type()
	var pointers int
	for pointer.Kind() == reflect.Ptr {
		pointer = pointer.Elem()
		pointers++
	}
	receiver := reflect.Zero(pointer)
	for range pointers {
		receiverPointer := reflect.New(receiver.Type())
		receiverPointer.Elem().Set(receiver)
		receiver = receiverPointer
	}
	return receiver
}

func newValueFromPointer(ctx context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	// this could be a nil pointer, let's make our own that we can set
	pointer := reflect.New(target.Type().Elem())
	// now we need to populate whatever the pointer is pointing to
	// we can't just use recursivelyPopulatePointer for this because we
	// want to check at every recursion depth whether we're implementing
	// the ValueSetter interface
	// newValue will still take care of recursing if we're pointing to a pointer
	pointed, err := newValue(ctx, flag, pointer.Elem())
	if err != nil {
		return target, err
	}
	// our new pointer doesn't come into the world settable
	// we need to make it settable by stuffing it in a pointer
	pointerPointer := reflect.New(pointer.Type())
	pointerPointer.Elem().Set(pointer)
	// our pointer is now settable, so stuff whatever we just created inside it
	pointerPointer.Elem().Elem().Set(pointed)
	// return the pointer we created
	return pointerPointer.Elem(), nil
}

func newValuesFromPointer(ctx context.Context, values FlagValues, target reflect.Value) (reflect.Value, error) {
	// this could be a nil pointer, let's make our own that we can set
	pointer := reflect.New(target.Type().Elem())
	// now we need to populate whatever the pointer is pointing to
	// we can't just use recursivelyPopulatePointer for this because we
	// want to check at every recursion depth whether we're implementing
	// the ValuesSetter interface
	// newValues will still take care of recursing if we're pointing to a pointer
	pointed, err := newValues(ctx, values, pointer.Elem())
	if err != nil {
		return target, err
	}
	// our new pointer doesn't come into the world settable
	// we need to make it settable by stuffing it in a pointer
	pointerPointer := reflect.New(pointer.Type())
	pointerPointer.Elem().Set(pointer)
	// our pointer is now settable, so stuff whatever we just created inside it
	pointerPointer.Elem().Elem().Set(pointed)
	// return the pointer we created
	return pointerPointer.Elem(), nil
}

func newSetFromPointer(ctx context.Context, flags FlagSet, target reflect.Value) (reflect.Value, error) {
	// this could be a nil pointer, let's make our own that we can set
	pointer := reflect.New(target.Type().Elem())
	// now we need to populate whatever the pointer is pointing to
	// we can't just use recursivelyPopulatePointer for this because we
	// want to check at every recursion depth whether we're implementing
	// the ValuesSetter interface
	// newValues will still take care of recursing if we're pointing to a pointer
	pointed, err := newSet(ctx, flags, pointer.Elem())
	if err != nil {
		return target, err
	}
	// our new pointer doesn't come into the world settable
	// we need to make it settable by stuffing it in a pointer
	pointerPointer := reflect.New(pointer.Type())
	pointerPointer.Elem().Set(pointer)
	// our pointer is now settable, so stuff whatever we just created inside it
	pointerPointer.Elem().Elem().Set(pointed)
	// return the pointer we created
	return pointerPointer.Elem(), nil
}
