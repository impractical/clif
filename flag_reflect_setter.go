package clif

import (
	"context"
	"reflect"
)

func newValueFromValueSetter(ctx context.Context, flag FlagValue, target reflect.Value) (reflect.Value, error) {
	methodName := "SetFromFlagValue"
	receiver := recursivelyPopulatePointer(ctx, target)
	method := receiver.MethodByName(methodName)
	if !method.IsValid() {
		return target, FlagSetterMethodInvalidError{
			Receiver: receiver.Interface(),
			Method:   methodName,
		}
	}
	results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(flag)})
	err := results[0].Interface()
	if err != nil {
		if e, ok := err.(error); ok {
			return target, e
		}
		return target, FlagSetterResponseError{
			Receiver: receiver.Interface(),
			Method:   methodName,
			Value:    err,
		}
	}
	return receiver, nil
}

func newValuesFromValuesSetter(ctx context.Context, values FlagValues, target reflect.Value) (reflect.Value, error) {
	methodName := "SetFromFlagValues"
	receiver := recursivelyPopulatePointer(ctx, target)
	method := receiver.MethodByName(methodName)
	if !method.IsValid() {
		return target, FlagSetterMethodInvalidError{
			Receiver: receiver.Interface(),
			Method:   methodName,
		}
	}
	results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(values)})
	err := results[0].Interface()
	if err != nil {
		if e, ok := err.(error); ok {
			return target, e
		}
		return target, FlagSetterResponseError{
			Receiver: receiver.Interface(),
			Method:   methodName,
			Value:    err,
		}
	}
	return receiver, nil
}

func newSetFromSetter(ctx context.Context, flags FlagSet, target reflect.Value) (reflect.Value, error) {
	methodName := "SetFromFlagSet"
	receiver := recursivelyPopulatePointer(ctx, target)
	method := receiver.MethodByName(methodName)
	if !method.IsValid() {
		return target, FlagSetterMethodInvalidError{
			Receiver: receiver.Interface(),
			Method:   methodName,
		}
	}
	results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(flags)})
	err := results[0].Interface()
	if err != nil {
		if e, ok := err.(error); ok {
			return target, e
		}
		return target, FlagSetterResponseError{
			Receiver: receiver.Interface(),
			Method:   methodName,
			Value:    err,
		}
	}
	return receiver, nil
}
