package peripheral

import (
	"github.com/Manu343726/cucaracha/pkg/logging"
)

// log returns the logger for the peripheral package.
// It uses hierarchical naming to create child loggers automatically.
func log() *logging.Logger {
	return logging.Get("cucaracha.hw.peripheral")
}
