// Package defaults provides embedded default configuration values for the
// Kion CLI application. It uses Go's embed functionality to include a YAML
// configuration file in the compiled binary, allowing for custom builds with
// predefined settings while maintaining the standard precedence of flags,
// environment variables, config files, and default values.
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
