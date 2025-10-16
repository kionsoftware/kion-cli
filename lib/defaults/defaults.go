// Package defaults provides access to the default configuration file embedded
// in the binary. This file is used to initialize the configuration file for
// organizations that want to distribute a binary with custom defaults.
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
