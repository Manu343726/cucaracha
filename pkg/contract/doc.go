package contract

// Package contract provides design-by-contract support for preconditions and postconditions.
//
// This package implements contract checking using generics, allowing developers to validate
// function inputs (preconditions), outputs (postconditions), and invariants at runtime.
//
// When a contract violation is detected, the logger will log the violation with a backtrace
// and panic, providing a clear error trail for debugging.
//
// Example usage:
//
//	func ProcessData(logger *logging.Logger, data []int) int {
//	    // Precondition: data must not be empty
//	    contract.Require(logger, len(data) > 0, "data must not be empty")
//
//	    // Precondition: all values must be positive
//	    contract.RequireEach[int](logger, data, func(v int) bool { return v > 0 },
//	        "all values must be positive")
//
//	    result := sum(data)
//
//	    // Postcondition: result must be greater than any individual element
//	    contract.Ensure(logger, result >= data[0], "result must be >= first element")
//
//	    // Postcondition: check result type constraint
//	    contract.EnsureValue[int](logger, result, func(r int) bool { return r > 0 },
//	        "result must be positive")
//
//	    return result
//	}
