package helper

import (
	"errors"
	"fmt"
	"os"

	"github.com/kionsoftware/kion-cli/lib/defaults"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"gopkg.in/yaml.v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Configuration                                                             //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// LoadConfig reads in the embedded configuration file as well as the users
// configuration yaml file located at `configFile`. Precedence is given to the
// users configuration file, so any values set there will override the embedded
// defaults.
func LoadConfig(filename string, config *structs.Configuration) error {
	// read embedded defaults into the config
	defaultConfig, err := defaults.GetDefaultConfig()
	if err == nil {
		// only try to parse if we successfully got the embedded config
		if err := yaml.Unmarshal(defaultConfig, &config); err != nil {
			return fmt.Errorf("failed to parse embedded configuration: %w", err)
		}
	}

	// try to read users config file
	data, err := os.ReadFile(filename)
	if err != nil {
		// if file doesn't exist, just use embedded defaults
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	// parse external config and override defaults
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return nil
}

// SaveConfig saves the entirety of the current config to the users config file.
func SaveConfig(filename string, config structs.Configuration) error {
	// marshal to yaml
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// write it out
	return os.WriteFile(filename, bytes, 0644)
}
