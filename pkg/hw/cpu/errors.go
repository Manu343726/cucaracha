package cpu

import "fmt"

type Error error

func makeError(err Error, message string, args ...interface{}) Error {
	return fmt.Errorf("%w: "+message, append([]any{err}, args...)...)
}
