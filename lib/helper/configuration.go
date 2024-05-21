package helper

import (
	"os"

	"github.com/kionsoftware/kion-cli/lib/structs"

	"gopkg.in/yaml.v2"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Configuration                                                             //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// LoadConfig reads in the configuration yaml file located at `configFile`.
func LoadConfig(filename string, config *structs.Configuration) error {
	// read in the file
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// unmarshal to config struct
	return yaml.Unmarshal(bytes, &config)
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

// SaveSession updates the session section only of the users config file.
func SaveSession(filename string, config structs.Configuration) error {
	// load in the current config file
	var newConfig structs.Configuration
	err := LoadConfig(filename, &newConfig)
	if err != nil {
		return err
	}

	// replace just the session
	newConfig.Session = config.Session

	// marshal to yaml
	bytes, err := yaml.Marshal(newConfig)
	if err != nil {
		return err
	}

	// write it out
	return os.WriteFile(filename, bytes, 0644)
}
