package cpu

import (
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// log returns the logger for the CPU package.
// It uses a child logger with hierarchical naming.
func log() *logging.Logger {
	return logging.Get("cucaracha.hw.cpu")
}
