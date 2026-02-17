package interpreter

import (
	"github.com/Manu343726/cucaracha/pkg/logging"
)

// log returns the logger for the interpreter package.
func log() *logging.Logger {
	return logging.Get("cucaracha.interpreter")
}
