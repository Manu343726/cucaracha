//go:build !windows
// +build !windows

package cpu

import (
	"os"
	"os/signal"
	"syscall"
)

// monitorResize uses SIGWINCH signal on Unix systems for efficient resize detection
func (ui *cliUI) monitorResize() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	defer signal.Stop(sigChan)

	for {
		select {
		case <-ui.resizeStop:
			return
		case <-sigChan:
			size := ui.GetTerminalSize()
			ui.notifyResizeHandlers(size)
		}
	}
}
