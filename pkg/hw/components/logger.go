package components

import (
	"github.com/Manu343726/cucaracha/pkg/logging"
)

// log returns the logger for the components package.
// It uses hierarchical naming to create child loggers automatically.
func log() *logging.Logger {
	return logging.Get("cucaracha.hw.components")
}
