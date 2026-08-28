package helper

import (
	"errors"
	"fmt"
	"os"

	"github.com/kionsoftware/kion-cli/lib/defaults"
	"github.com/kionsoftware/kion-cli/lib/structs"

	"gopkg.in/yaml.v3"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Configuration                                                             //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// LoadConfigStruct reads in the embedded configuration file as well as the users
// configuration yaml file located at `configFile`. Precedence is given to the
// users configuration file, so any values set there will override the embedded
// defaults.
func LoadConfigStruct(filename string, config *structs.Configuration) error {
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

// LoadConfig reads the configuration file from disk and parses it into a *yaml.Node tree.
// This method is preferable to LoadConfigStruct when you want to modify the configuration
// and write it back to disk, since it preserves comments and formatting.
func LoadConfig(configPath string) (*yaml.Node, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file %s: %w", configPath, err)
	}

	var configData yaml.Node
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("error parsing configuration file %s: %w", configPath, err)
	}

	return &configData, nil
}

// SaveConfig marshals a *yaml.Node tree back to yaml and writes it to disk.
func SaveConfig(configPath string, configData *yaml.Node) error {
	out, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("error marshaling configuration: %w", err)
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return fmt.Errorf("error writing configuration file %s: %w", configPath, err)
	}

	return nil
}

// UpdateConfigField sets the value at fieldPath in the YAML file at configPath,
// preserving existing comments and formatting. It loads the file, applies the
// update via setField, and writes the result back to disk in a single call.
func UpdateConfigField(configPath string, fieldPath []string, value string) error {
	configData, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	err = setField(configData, fieldPath, value)
	if err != nil {
		return err
	}
	err = SaveConfig(configPath, configData)
	if err != nil {
		return err
	}
	return nil
}

// DeleteConfigField removes the key at fieldPath — along with its entire
// value subtree — from the YAML file at configPath. It loads the file,
// removes the field via deleteField, and writes the result back to disk in
// a single call.
func DeleteConfigField(configPath string, fieldPath []string) error {
	configData, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	err = deleteField(configData, fieldPath)
	if err != nil {
		return err
	}
	err = SaveConfig(configPath, configData)
	if err != nil {
		return err
	}
	return nil
}

// deleteField removes a key (and its entire value subtree, including any
// nested children) from a mapping node at the given key path. It walks to
// the parent mapping of the final key, then removes both the key node and
// its value node from that parent's Content slice.
func deleteField(root *yaml.Node, path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("empty key path")
	}

	// root.Content[0] is the actual document mapping
	node := root.Content[0]

	// Walk to the parent mapping of the key we want to delete.
	for _, key := range path[:len(path)-1] {
		found := false
		for j := 0; j < len(node.Content)-1; j += 2 {
			k := node.Content[j]
			v := node.Content[j+1]
			if k.Value == key {
				found = true
				node = v
				break
			}
		}
		if !found {
			return fmt.Errorf("key path not found at %q", key)
		}
	}

	// node is now the parent mapping; remove the final key/value pair.
	lastKey := path[len(path)-1]
	for j := 0; j < len(node.Content)-1; j += 2 {
		if node.Content[j].Value == lastKey {
			// Cut out both the key node (j) and value node (j+1).
			node.Content = append(node.Content[:j], node.Content[j+2:]...)
			return nil
		}
	}

	return fmt.Errorf("key path not found at %q", lastKey)
}

// setField walks a mapping node by key path and updates the scalar value at the end.
func setField(root *yaml.Node, path []string, value string) error {
	// root.Content[0] is the actual document mapping
	node := root.Content[0]

	for i, key := range path {
		found := false
		for j := 0; j < len(node.Content)-1; j += 2 {
			k := node.Content[j]
			v := node.Content[j+1]
			if k.Value == key {
				found = true
				if i == len(path)-1 {
					v.Value = value
				} else {
					node = v
				}
				break
			}
		}
		if !found {
			return fmt.Errorf("key path not found at %q", key)
		}
	}
	return nil
}
