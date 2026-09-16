package config

import (
	"fmt"
	"os"
	"strings"

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
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err 
	}

	var cfg OrbitConfig

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err 
	}

	// Validation
	for name, task := range cfg.Tasks {
		if strings.TrimSpace(task.Command) == "" {
			return nil, fmt.Errorf("task %q has an empty command", name)
		}
		if strings.ContainsAny(name, " \t\n\r") {
			return nil, fmt.Errorf("task name %q contains invalid whitespace characters", name)
		}
	}

	return &cfg, nil
}
