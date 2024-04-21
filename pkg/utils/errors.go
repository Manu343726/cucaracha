package utils

import (
	"fmt"
)

func MakeError(err error, detailsBody string, args ...any) error {
	return fmt.Errorf("%w: "+detailsBody, append([]any{err}, args...))
}
