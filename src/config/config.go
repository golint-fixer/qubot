package config

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

// Config is the conguration of Qubot.
type Config struct {
	Database *DatabaseConfig
	Slack    *SlackConfig
	Redmine  *RedmineConfig
}

// DatabaseConfig is the database configuration.
type DatabaseConfig struct {
	Location string
}

// SlackConfig holds the configuration parameters to access Slack.
type SlackConfig struct {
	Key string
}

// RedmineConfig is the configuration of Redmine.
type RedmineConfig struct {
	URL  string
	Key  string
	User string
}

// DefaultConfig is the default configuration
var DefaultConfig *Config

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

// Load loads the CLI configuration from ".qubotrc" files.
func Load(path string) (*Config, error) {
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
	var result Config
	if err := hcl.DecodeObject(&result, obj); err != nil {
		return nil, err
	}

	return &result, nil
}
