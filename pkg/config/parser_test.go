package config

import (
	"os"
	"testing"
)

func TestParseConfig_Valid(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "orbit.yaml")
	defer os.Remove(tmpFile.Name())

	yamlContent := `
tasks:
  build:
    command: "go build"
    depends_on: []
`
	tmpFile.Write([]byte(yamlContent))
	tmpFile.Close()

	cfg, err := ParseConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(cfg.Tasks))
	}
}

func TestParseConfig_EmptyCommand(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "orbit.yaml")
	defer os.Remove(tmpFile.Name())

	yamlContent := `
tasks:
  build:
    command: ""
    depends_on: []
`
	tmpFile.Write([]byte(yamlContent))
	tmpFile.Close()

	_, err := ParseConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
}

func TestParseConfig_InvalidName(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "orbit.yaml")
	defer os.Remove(tmpFile.Name())

	yamlContent := `
tasks:
  "bad name":
    command: "echo hi"
    depends_on: []
`
	tmpFile.Write([]byte(yamlContent))
	tmpFile.Close()

	_, err := ParseConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for invalid name, got nil")
	}
}

func TestParseConfig_MalformedYAML(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "orbit.yaml")
	defer os.Remove(tmpFile.Name())

	yamlContent := `
tasks:
  build:
    command "go build"
`
	tmpFile.Write([]byte(yamlContent))
	tmpFile.Close()

	_, err := ParseConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestParseConfig_ZeroTasks(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "orbit.yaml")
	defer os.Remove(tmpFile.Name())

	yamlContent := `
tasks: {}
`
	tmpFile.Write([]byte(yamlContent))
	tmpFile.Close()

	_, err := ParseConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for zero tasks, got nil")
	}
}
