package templates

import (
	"github.com/Manu343726/cucaracha/pkg/utils/logging"
)

// log returns the logger for the LLVM templates package.
// It uses hierarchical naming to create child loggers automatically.
func log() *logging.Logger {
	return logging.Get("cucaracha.llvm.templates")
}
