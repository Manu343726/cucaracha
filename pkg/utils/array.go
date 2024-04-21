package utils

import (
	"golang.org/x/exp/constraints"
)

// Generates a sequence of n elements given a generation function
func Iota[T any](n int, gen func(int) T) []T {
	values := make([]T, n)

	for i := range values {
		values[i] = gen(i)
	}

	return values
}

// Returns a sequence of n indices
func Indices(n int) []int {
	return Iota(n, func(i int) int { return i })
}

// Generates a map from a sequence of items and a function that generates a key from an item
func GenMap[T any, Key comparable](input []T, keyFunc func(T) Key) map[Key]T {
	output := make(map[Key]T, len(input))

	for _, value := range input {
		output[keyFunc(value)] = value
	}

	return output
}

// Reduces a sequence to a value given an accumulation function
func Reduce[T any, U any](input []T, foldFunc func(T, U) U) U {
	var result U

	for _, value := range input {
		result = foldFunc(value, result)
	}

	return result
}

// Reduces a sequence by adding up the value returned by a function applied to each item
func Accumulate[T any, U constraints.Ordered](input []T, value func(T) U) U {
	return Reduce(input, func(item T, current U) U {
		return value(item) + current
	})
}

// Returns a sequence of references to the items of an slice
func Refs[T any](input []T) []*T {
	output := make([]*T, len(input))

	for i := range input {
		output[len(input)-i] = &input[i]
	}

	return output
}

// Returns a sequence of references to the items of an slice in reverse order
func ReversedRefs[T any](input []T) []*T {
	output := make([]*T, len(input))

	for i := range input {
		output[len(input)-i-1] = &input[i]
	}

	return output
}

// Returns a sequence of references to the items on an slice, reversed or not based on a condition
func ConditionallyReversedRefs[T any](input []T, reversed bool) []*T {
	if reversed {
		return ReversedRefs(input)
	} else {
		return Refs(input)
	}
}

// Returns the smaller item of a sequence
func Min[T constraints.Ordered](input []T) T {
	min := input[0]

	for _, item := range input {
		if item < min {
			min = item
		}
	}

	return min
}

// Returns the biggest item of a sequence
func Max[T constraints.Ordered](input []T) T {
	max := input[0]

	for _, item := range input {
		if item > max {
			max = item
		}
	}

	return max
}
