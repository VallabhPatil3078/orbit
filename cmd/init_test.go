package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmd(t *testing.T) {
	// Create a temporary directory to act as our repository
	tmpDir, err := os.MkdirTemp("", "orbit-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Remember the original working directory and restore it later
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change into the temp directory
	os.Chdir(tmpDir)

	// Create a fake .git/hooks directory so init doesn't skip hook installation
	os.MkdirAll(filepath.Join(".git", "hooks"), 0755)

	// Run the init command
	initCmd.Run(initCmd, []string{})

	// Verify orbit.yaml was created
	if _, err := os.Stat("orbit.yaml"); os.IsNotExist(err) {
		t.Error("orbit.yaml was not created")
	}

	// Verify pre-commit hook was created
	hookPath := filepath.Join(".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("pre-commit hook was not created")
	}
}
