package utils

import (
	"fmt"
	"reflect"
)

// Returns the value of an object member by name.
// If the member is a method it is assumed that it
// has no paramters and gets called to return the value.
func Member(name string, object any) (any, error) {
	v := reflect.ValueOf(object)

	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object is not a struct, cannot reference field '%v'", name)
	}

	if result := v.FieldByName(name); result.IsValid() {
		return result.Interface(), nil
	} else if result := v.MethodByName(name); result.IsValid() {
		return result.Call(nil)[0].Interface(), nil
	} else if result := v.Addr().MethodByName(name); result.IsValid() {
		return result.Call(nil)[0].Interface(), nil
	} else {
		return nil, fmt.Errorf("struct '%v' has no field or method named '%v'", v.Type().Name(), name)
	}
}

// Maps a sequence of objects into a sequence of values of a given member of each item of the input sequence
func MapMember(name string, items []any) ([]any, error) {
	result := make([]any, len(items))

	for i, object := range items {
		value, err := Member(name, object)

		if err != nil {
			return nil, err
		}

		result[i] = value
	}

	return result, nil
}
