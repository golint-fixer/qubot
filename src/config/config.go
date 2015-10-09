package config

import (
	"fmt"
	"io/ioutil"

	"qubot"

	"github.com/hashicorp/hcl"
)

// DefaultConfig is the default configuration
var DefaultConfig *qubot.Config

// File returns the default path to the configuration file.
//
// On Unix-like systems this is the ".qubotrc" file in the home directory.
// On Windows, this is the "qubot.rc" file in the application data
// directory.
func File() (string, error) {
	return configFile()
}

// Dir returns the configuration directory for Qubot.
func Dir() (string, error) {
	return configDir()
}

// Load loads the configuration from ".qubotrc" files.
func Load(path string) (*qubot.Config, error) {
	// Read the HCL file and prepare for parsing
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %s", path, err)
	}

	// Parse it
	obj, err := hcl.Parse(string(d))
	if err != nil {
		return nil, fmt.Errorf(
			"Error parsing %s: %s", path, err)
	}

	// Build up the result
	var result qubot.Config
	if err := hcl.DecodeObject(&result, obj); err != nil {
		return nil, err
	}

	return &result, nil
}
