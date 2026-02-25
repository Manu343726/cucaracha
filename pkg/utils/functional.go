package utils

import "fmt"

// Converts a function taking T and returning (R, error) into a function taking any and returning (any, error)
func UntypeFunction[T any, R any](f func(T) (R, error)) func(any) (any, error) {
	return func(arg any) (any, error) {
		typedArg, ok := arg.(T)
		if !ok {
			return nil, fmt.Errorf("invalid argument type: expected %T", *new(T))
		}
		result, err := f(typedArg)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}
