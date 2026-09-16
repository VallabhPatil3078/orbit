package config

import (
	"os"
	"strings"
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
	if !strings.Contains(err.Error(), "empty command") {
		t.Errorf("expected error to mention empty command, got: %v", err)
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
	if !strings.Contains(err.Error(), "invalid whitespace characters") {
		t.Errorf("expected error to mention invalid whitespace, got: %v", err)
	}
}
