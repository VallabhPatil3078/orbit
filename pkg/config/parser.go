package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TaskConfig represents a single task definition from the YAML file.
type TaskConfig struct {
	Command   string   `yaml:"command"`
	DependsOn []string `yaml:"depends_on"`
}

// OrbitConfig represents the entire orbit.yaml file structure.
type OrbitConfig struct {
	Tasks map[string]TaskConfig `yaml:"tasks"`
}

// ParseConfig reads a YAML file and unmarshals it into the OrbitConfig struct.
func ParseConfig(filepath string) (*OrbitConfig, error) {
	// Read the raw bytes from the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err // Return the error if the file doesn't exist or is unreadable
	}

	// Create an empty instance of our struct
	var cfg OrbitConfig

	// Unmarshal converts the raw YAML bytes into our Go struct
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err // Return an error if the YAML format is invalid
	}

	return &cfg, nil
}
