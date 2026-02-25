package reflect

import (
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// log returns the logger for the reflect package.
func log() *logging.Logger {
	return logging.Get("cucaracha.reflect")
}
