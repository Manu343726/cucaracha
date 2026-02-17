package contract

import (
	"reflect"

	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// Base is an embeddable type that provides logging and contract checking capabilities.
// Types that embed Base can use log() to get a logger and call contract methods
// without explicitly passing a logger.
//
// Example:
//
//	type Service struct {
//	    contract.Base
//	}
//
//	func (s *Service) Process(data []int) int {
//	    s.NotNil(data)
//	    s.log().Debug("processing data", slog.Int("count", len(data)))
//	    result := s.computeResult(data)
//	    s.Ensure(result > 0, "result must be positive")
//	    return result
//	}
type Base struct {
	logger *logging.Logger
}

func NewBase(logger *logging.Logger) Base {
	return Base{logger: logger}
}

// log returns the embedded logger for direct logging calls.
func (b *Base) Log() *logging.Logger {
	return b.logger
}

// Require checks a precondition and panics if it's false.
//
//	s.Require(x > 0, "x must be positive")
//	s.Require(x > 0, "x=%d must be positive", x)
func (b *Base) Require(condition bool, fmtAndArgs ...any) {
	Require(b.logger, condition, fmtAndArgs...)
}

// Ensure checks a postcondition and panics if it's false.
//
//	s.Ensure(result > 0, "result must be positive")
//	s.Ensure(result > 0, "result=%d but expected >0", result)
func (b *Base) Ensure(condition bool, fmtAndArgs ...any) {
	Ensure(b.logger, condition, fmtAndArgs...)
}

// Invariant checks an invariant condition and panics if it's false.
//
//	s.Invariant(active, "must be active")
//	s.Invariant(active, "state=%s invalid", state)
func (b *Base) Invariant(condition bool, fmtAndArgs ...any) {
	Invariant(b.logger, condition, fmtAndArgs...)
}

// RequireExpr checks a precondition using an expression and panics if it's false.
func (b *Base) RequireExpr(expr *Expr, fmtAndArgs ...any) {
	RequireExpr(b.logger, expr, fmtAndArgs...)
}

// EnsureExpr checks a postcondition using an expression and panics if it's false.
func (b *Base) EnsureExpr(expr *Expr, fmtAndArgs ...any) {
	EnsureExpr(b.logger, expr, fmtAndArgs...)
}

// InvariantExpr checks an invariant condition using an expression and panics if it's false.
func (b *Base) InvariantExpr(expr *Expr, fmtAndArgs ...any) {
	InvariantExpr(b.logger, expr, fmtAndArgs...)
}

// RequireThat checks a precondition with automatic variable capture from a struct or map.
func (b *Base) RequireThat(condition bool, args ...any) {
	RequireThat(b.logger, condition, args...)
}

// EnsureThat checks a postcondition with automatic variable capture from a struct or map.
func (b *Base) EnsureThat(condition bool, args ...any) {
	EnsureThat(b.logger, condition, args...)
}

// InvariantThat checks an invariant with automatic variable capture from a struct or map.
func (b *Base) InvariantThat(condition bool, args ...any) {
	InvariantThat(b.logger, condition, args...)
}

// RequireValue checks a precondition on a value using a validator function.
func (b *Base) RequireValue(value interface{}, validator func(interface{}) bool, fmtAndArgs ...any) {
	Require(b.logger, validator(value), fmtAndArgs...)
}

// EnsureValue checks a postcondition on a value using a validator function.
func (b *Base) EnsureValue(value interface{}, validator func(interface{}) bool, fmtAndArgs ...any) {
	Ensure(b.logger, validator(value), fmtAndArgs...)
}

// InvariantValue checks an invariant on a value using a validator function.
func (b *Base) InvariantValue(value interface{}, validator func(interface{}) bool, fmtAndArgs ...any) {
	Invariant(b.logger, validator(value), fmtAndArgs...)
}

// ExpectThat is a simple contract that logs a failure message without panicking.
func (b *Base) ExpectThat(condition bool) {
	ExpectThat(b.logger, condition)
}

// RequireEach checks a precondition for each element in a collection.
func (b *Base) RequireEach(values interface{}, validator func(interface{}) bool, fmtAndArgs ...any) {
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		b.logger.Panic("RequireEach requires a slice")
		return
	}

	for i := 0; i < rv.Len(); i++ {
		Require(b.logger, validator(rv.Index(i).Interface()), fmtAndArgs...)
	}
}

// EnsureEach checks a postcondition for each element in a collection.
func (b *Base) EnsureEach(values interface{}, validator func(interface{}) bool, fmtAndArgs ...any) {
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		b.logger.Panic("EnsureEach requires a slice")
		return
	}

	for i := 0; i < rv.Len(); i++ {
		Ensure(b.logger, validator(rv.Index(i).Interface()), fmtAndArgs...)
	}
}

// InvariantEach checks an invariant for each element in a collection.
func (b *Base) InvariantEach(values interface{}, validator func(interface{}) bool, fmtAndArgs ...any) {
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		b.logger.Panic("InvariantEach requires a slice")
		return
	}

	for i := 0; i < rv.Len(); i++ {
		Invariant(b.logger, validator(rv.Index(i).Interface()), fmtAndArgs...)
	}
}

// NotNil checks that a value is not nil and panics if it is.
func (b *Base) NotNil(value interface{}, fmtAndArgs ...any) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr && rv.Kind() != reflect.Interface && rv.Kind() != reflect.Slice && rv.Kind() != reflect.Map {
		return
	}
	Require(b.logger, !rv.IsNil(), fmtAndArgs...)
}

// NotEmpty checks that a string is not empty and panics if it is.
func (b *Base) NotEmpty(value string, fmtAndArgs ...any) {
	NotEmpty(b.logger, value, fmtAndArgs...)
}

// NotEmptySlice checks that a slice is not empty and panics if it is.
func (b *Base) NotEmptySlice(values interface{}, fmtAndArgs ...any) {
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		b.logger.Panic("NotEmptySlice requires a slice")
		return
	}
	Require(b.logger, rv.Len() > 0, fmtAndArgs...)
}

// RequireRange checks that an integer is within a range and panics if it's not.
func (b *Base) RequireRange(value, min, max int, fmtAndArgs ...any) {
	RangeInt(b.logger, value, min, max, fmtAndArgs...)
}
