package hw

import (
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// log returns the logger for the hardware package.
func log() *logging.Logger {
	return logging.Get("cucaracha.hw")
}
