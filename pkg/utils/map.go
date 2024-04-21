package utils

// Generates a sequence constructed by applying a function to all elements of a given input sequence
func Map[T any, U any](input []T, mapFunction func(T) U) []U {
	output := make([]U, len(input))

	for i := range input {
		output[i] = mapFunction(input[i])
	}

	return output
}

// Generates a new Map NewKey -> NewValue from a given map Key -> Value and a transformation function (Key, Value) -> (NewKey, NewValue)
func MapMap[Key comparable, Value comparable, NewKey comparable, NewValue comparable](input map[Key]Value, mapFunction func(Key, Value) (NewKey, NewValue)) map[NewKey]NewValue {
	output := make(map[NewKey]NewValue, len(input))

	for key, value := range input {
		newKey, newValue := mapFunction(key, value)
		output[newKey] = newValue
	}

	return output
}

// Converts a Key -> Value map into a Value -> Key map
func InvertedMap[Key comparable, Value comparable](input map[Key]Value) map[Value]Key {
	return MapMap(input, func(key Key, value Value) (Value, Key) {
		return value, key
	})
}

// Returns an array with all the keys of a map
func Keys[Key comparable, Value comparable](input map[Key]Value) []Key {
	keys := make([]Key, 0, len(input))

	for key := range input {
		keys = append(keys, key)
	}

	return keys
}

// Returns an array with all the values of a map
func Values[Key comparable, Value comparable](input map[Key]Value) []Value {
	values := make([]Value, 0, len(input))

	for _, value := range input {
		values = append(values, value)
	}

	return values
}

// Returns an array of pairs (Key, Value) from a given map Key -> Value
func ZipMap[Key comparable, Value comparable](input map[Key]Value) []Pair[Key, Value] {
	pairs := make([]Pair[Key, Value], 0, len(input))

	for key, value := range input {
		pairs = append(pairs, MakePair(key, value))
	}

	return pairs
}
