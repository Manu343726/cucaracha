package contract

import (
	"fmt"
	"log/slog"

	"github.com/Manu343726/cucaracha/pkg/logging"
)

// FormatErrorMessage formats an error message like fmt.Errorf.
// If no format arguments provided, uses the first string as message.
func FormatErrorMessage(defaultMsg string, fmtAndArgs ...any) string {
	if len(fmtAndArgs) == 0 {
		return defaultMsg
	}
	if len(fmtAndArgs) == 1 {
		if s, ok := fmtAndArgs[0].(string); ok {
			return s
		}
	}
	if len(fmtAndArgs) > 0 {
		if format, ok := fmtAndArgs[0].(string); ok {
			return fmt.Sprintf(format, fmtAndArgs[1:]...)
		}
	}
	return defaultMsg
}

// Require checks a precondition and panics if it's false.
// Preconditions should be checked at the beginning of a function to validate inputs.
//
// Example:
//
//	func Divide(logger *logging.Logger, a, b int) int {
//	    contract.Require(logger, b != 0, "divisor must not be zero")
//	    return a / b
//	}
func Require(logger *logging.Logger, condition bool, fmtAndArgs ...any) {
	if !condition {
		msg := FormatErrorMessage("precondition failed", fmtAndArgs...)
		logger.Panic(msg)
	}
}

// RequireValue checks a precondition using a validator function and panics if validation fails.
// The validator function receives the value and should return true if the condition is met.
//
// Example:
//
//	func OpenFile(logger *logging.Logger, path string) *os.File {
//	    contract.RequireValue[string](logger, path, func(p string) bool {
//	        return len(p) > 0
//	    }, "file path must not be empty")
//	    // ...
//	}
func RequireValue[T any](logger *logging.Logger, value T, validator func(T) bool, fmtAndArgs ...any) {
	Require(logger, validator(value), fmtAndArgs...)
}

// RequireEach checks a precondition for each element in a slice and panics on first failure.
// This is useful for validating collections of inputs.
//
// Example:
//
//	func ProcessIDs(logger *logging.Logger, ids []int) {
//	    contract.RequireEach[int](logger, ids, func(id int) bool {
//	        return id > 0
//	    }, "all IDs must be positive")
//	}
func RequireEach[T any](logger *logging.Logger, values []T, validator func(T) bool, fmtAndArgs ...any) {
	for i, v := range values {
		if !validator(v) {
			msg := FormatErrorMessage("precondition failed", fmtAndArgs...)
			logger.Panic(msg, slog.Int("index", i))
		}
	}
}

// Ensure checks a postcondition and panics if it's false.
// Postconditions should be checked at the end of a function to validate return values.
//
// Example:
//
//	func Divide(logger *logging.Logger, a, b int) int {
//	    result := a / b
//	    contract.Ensure(logger, result >= 0, "result must be non-negative")
//	    return result
//	}
func Ensure(logger *logging.Logger, condition bool, fmtAndArgs ...any) {
	if !condition {
		msg := FormatErrorMessage("postcondition failed", fmtAndArgs...)
		logger.Panic(msg)
	}
}

// EnsureValue checks a postcondition using a validator function and panics if validation fails.
// The validator function receives the return value and should return true if the condition is met.
//
// Example:
//
//	func Divide(logger *logging.Logger, a, b int) int {
//	    result := a / b
//	    contract.EnsureValue[int](logger, result, func(r int) bool {
//	        return r >= 0
//	    }, "result must be non-negative")
//	    return result
//	}
func EnsureValue[T any](logger *logging.Logger, value T, validator func(T) bool, fmtAndArgs ...any) {
	Ensure(logger, validator(value), fmtAndArgs...)
}

// Invariant checks an invariant condition and panics if it's false.
// Invariants should be used to verify the consistency of an object's internal state.
//
// Example:
//
//	func (acc *Account) Deposit(logger *logging.Logger, amount int) {
//	    contract.Require(logger, amount > 0, "amount must be positive")
//	    acc.balance += amount
//	    contract.Invariant(logger, acc.balance >= 0, "balance must never be negative")
//	}
func Invariant(logger *logging.Logger, condition bool, fmtAndArgs ...any) {
	if !condition {
		msg := FormatErrorMessage("invariant violated", fmtAndArgs...)
		logger.Panic(msg)
	}
}

// InvariantEach checks an invariant for each element in a slice and panics on first failure.
//
// Example:
//
//	func (list *LinkedList) CheckInvariants(logger *logging.Logger) {
//	    contract.InvariantEach[*Node](logger, list.nodes, func(n *Node) bool {
//	        return n != nil
//	    }, "all nodes must be non-nil")
//	}
func InvariantEach[T any](logger *logging.Logger, values []T, validator func(T) bool, fmtAndArgs ...any) {
	for i, v := range values {
		if !validator(v) {
			msg := FormatErrorMessage("invariant violated", fmtAndArgs...)
			logger.Panic(msg, slog.Int("index", i))
		}
	}
}

// RangeInt checks that an integer is within a range and panics if it's not.
// This is a convenience function for common integer range checks.
//
// Example:
//
//	func SetAge(logger *logging.Logger, age int) {
//	    contract.RangeInt(logger, age, 0, 120, "age must be between 0 and 120")
//	}
func RangeInt(logger *logging.Logger, value, min, max int, fmtAndArgs ...any) {
	if value < min || value > max {
		msg := FormatErrorMessage("value out of range", fmtAndArgs...)
		logger.Panic(msg)
	}
}

// NotNil panics if the value is nil.
func NotNil[T any](logger *logging.Logger, value *T, fmtAndArgs ...any) {
	if value == nil {
		msg := FormatErrorMessage("value must not be nil", fmtAndArgs...)
		logger.Panic(msg)
	}
}

// NotEmpty panics if the string is empty.
func NotEmpty(logger *logging.Logger, value string, fmtAndArgs ...any) {
	if value == "" {
		msg := FormatErrorMessage("string must not be empty", fmtAndArgs...)
		logger.Panic(msg)
	}
}

// NotEmptySlice panics if the slice is empty.
func NotEmptySlice[T any](logger *logging.Logger, value []T, fmtAndArgs ...any) {
	if len(value) == 0 {
		msg := FormatErrorMessage("slice must not be empty", fmtAndArgs...)
		logger.Panic(msg)
	}
}
