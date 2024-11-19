package clif

import (
	"context"
	"errors"
	"reflect"
	"slices"
)

func newSet(ctx context.Context, flags FlagSet, target reflect.Value) (reflect.Value, error) {
	if target.Type().Implements(reflect.TypeOf((*FlagSetSetter)(nil)).Elem()) {
		return newSetFromSetter(ctx, flags, target)
	}

	// if we're pointing to a pointer, dereference it and try again on
	// whatever the pointer is pointing to
	if target.Kind() == reflect.Ptr {
		return newSetFromPointer(ctx, flags, target)
	}

	if target.Kind() == reflect.Map {
		return newSetFromMap(ctx, flags, target)
	}

	if target.Kind() != reflect.Struct {
		return target, InvalidKindError{
			Value: target.Interface(),
			Kind:  target.Kind(),
		}
	}

	// check that all flags have a field defined
	flagKeys := make(map[string]struct{}, len(flags))
	for key := range flags {
		flagKeys[key] = struct{}{}
	}
	for _, key := range getAllStructFlagKeys(target.Type()) {
		delete(flagKeys, key)
	}
	var errs error
	for key := range flagKeys {
		errs = errors.Join(errs, FlagKeyNotInStructError{
			Key:        key,
			StructType: target.Type().String(),
		})
	}
	if errs != nil {
		return target, errs
	}

	targetFields, err := getStructTags(ctx, target)
	if err != nil {
		return target, err
	}

	results := reflect.New(target.Type()).Elem()
	unusedFlags := FlagSet{}
	for key, values := range flags {
		unusedFlags[key] = slices.Clone(values)
	}
	for field, structFieldPos := range targetFields {
		structField := results.Field(structFieldPos)
		flag, ok := flags[field]
		var values reflect.Value
		if !ok {
			values = reflect.Zero(structField.Type())
		} else {
			values, err = newValues(ctx, flag, structField)
			if err != nil {
				return target, err
			}
			delete(unusedFlags, field)
		}
		structField.Set(values)
	}

	// recursively apply flags to any anonymous structs embedded in this
	// struct, too
	for fieldNum := range results.Type().NumField() {
		typeField := results.Type().Field(fieldNum)
		if !typeField.Anonymous {
			continue
		}
		if typeField.Type.Kind() != reflect.Struct {
			continue
		}
		field := results.Field(fieldNum)
		val, err := newSet(ctx, unusedFlags, field)
		if err != nil {
			return target, err
		}
		field.Set(val)
	}

	return results, nil
}

// getStructTags returns a map of flag keys to the position in
// the struct `in` that the flag values should be reflected into.
func getStructTags(_ context.Context, value reflect.Value) (map[string]int, error) {
	tags := map[string]int{}
	for fieldNum := range value.Type().NumField() {
		typeField := value.Type().Field(fieldNum)
		if typeField.PkgPath != "" {
			// skip unexported fields
			continue
		}
		if typeField.Anonymous {
			// skip anonymous structs embedded; we'll handle those separately
			continue
		}
		tag := typeField.Tag.Get("flag")
		if tag == "-" {
			// skip explicitly excluded fields
			continue
		}
		if tag == "" {
			return nil, StructFieldMissingFlagTagError{
				Struct: value.Interface(),
				Field:  typeField.Name,
			}
		}
		tags[tag] = fieldNum
	}
	return tags, nil
}

func getAllStructFlagKeys(structType reflect.Type) []string {
	var results []string
	for fieldNum := range structType.NumField() {
		typeField := structType.Field(fieldNum)
		if typeField.PkgPath != "" {
			// skip unexported fields
			continue
		}
		if typeField.Anonymous && typeField.Type.Kind() == reflect.Struct {
			results = append(results, getAllStructFlagKeys(typeField.Type)...)
		} else if typeField.Anonymous {
			// skip anonymous non-struct fields
			continue
		}
		tag := typeField.Tag.Get("flag")
		if tag == "-" {
			// skip explicitly excluded fields
			continue
		}
		if tag == "" {
			// skip fields without a tag, we'll throw an error
			// about them later
			continue
		}
		results = append(results, tag)
	}
	return results
}
