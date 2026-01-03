//go:build windows
// +build windows

package cpu

import (
	"time"
)

// monitorResize uses polling on Windows since there's no SIGWINCH equivalent
func (ui *cliUI) monitorResize() {
	lastSize := ui.GetTerminalSize()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ui.resizeStop:
			return
		case <-ticker.C:
			currentSize := ui.GetTerminalSize()
			if currentSize.Width != lastSize.Width || currentSize.Height != lastSize.Height {
				lastSize = currentSize
				ui.notifyResizeHandlers(currentSize)
			}
		}
	}
}
