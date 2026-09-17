package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type TaskConfig struct {
	Command      string   `yaml:"command"`
	WorkingDir   string   `yaml:"working_dir"`
	DependsOn    []string `yaml:"depends_on"`
	TriggerPaths []string `yaml:"trigger_paths"`
	IgnorePaths  []string `yaml:"ignore_paths"`
}

type OrbitConfig struct {
	Tasks map[string]TaskConfig `yaml:"tasks"`
}

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

	if len(cfg.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks defined in orbit.yaml")
	}

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

