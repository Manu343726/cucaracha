package widgets

import "github.com/Manu343726/cucaracha/pkg/logging"

func log() *logging.Logger {
	return logging.Get("cucaracha.ui.tui.widgets")
}
