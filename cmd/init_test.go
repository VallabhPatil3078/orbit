package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCmd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "orbit-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	os.MkdirAll(filepath.Join(".git", "hooks"), 0755)

	force = false
	initCmd.Run(initCmd, []string{})

	if _, err := os.Stat("orbit.yaml"); os.IsNotExist(err) {
		t.Error("orbit.yaml was not created")
	}

	hookPath := filepath.Join(".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("pre-commit hook was not created")
	}

	// Test overwriting without force
	os.WriteFile("orbit.yaml", []byte("old content"), 0644)
	initCmd.Run(initCmd, []string{})
	content, _ := os.ReadFile("orbit.yaml")
	if string(content) != "old content" {
		t.Error("orbit.yaml was overwritten without --force")
	}

	// Test hook backup
	os.Remove("orbit.yaml")
	os.WriteFile(hookPath, []byte("echo 'other hook'"), 0755)
	initCmd.Run(initCmd, []string{})
	if _, err := os.Stat(hookPath + ".orbit-backup"); os.IsNotExist(err) {
		t.Error("Existing hook was not backed up")
	}

	// Test force
	force = true
	os.WriteFile("orbit.yaml", []byte("old content"), 0644)
	initCmd.Run(initCmd, []string{})
	content, _ = os.ReadFile("orbit.yaml")
	if strings.Contains(string(content), "old content") {
		t.Error("orbit.yaml was not overwritten with --force")
	}
}
