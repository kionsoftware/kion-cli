package defaults

import (
	"embed"
)

//go:embed defaults.yml
var ConfigFS embed.FS

// GetDefaultConfig returns the default configuration.
func GetDefaultConfig() ([]byte, error) {
	return ConfigFS.ReadFile("defaults.yml")
}
