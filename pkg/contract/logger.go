package contract

import (
	"github.com/Manu343726/cucaracha/pkg/logging"
)

// log returns the logger for the contract package.
func log() *logging.Logger {
	return logging.Get("cucaracha.contract")
}
