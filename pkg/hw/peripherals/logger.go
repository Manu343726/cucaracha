package peripherals

import (
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// log returns the logger for the peripherals package.
// It uses hierarchical naming to create child loggers automatically.
func log() *logging.Logger {
	return logging.Get("cucaracha.hw.peripherals")
}
