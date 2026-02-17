package debugger

import (
	_ "embed"

	"github.com/Manu343726/cucaracha/pkg/system"
)

//go:embed default_system.yaml
var DefaultSystemConfigYAML string

// DefaultSystemConfig returns the default system configuration for the debugger
func DefaultSystemConfig() (*system.SystemConfig, error) {
	return system.Parse([]byte(DefaultSystemConfigYAML))
}
