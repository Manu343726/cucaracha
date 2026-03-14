package disassembly

import (
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// log returns the logger for the disassembly package.
func log() *logging.Logger {
	return logging.Get("cucaracha.ui.disassembly")
}
