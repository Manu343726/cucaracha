package ui

import (
	"github.com/Manu343726/cucaracha/pkg/logging"
)

// log returns the logger for the UI package.
func log() *logging.Logger {
	return logging.Get("cucaracha.ui")
}
