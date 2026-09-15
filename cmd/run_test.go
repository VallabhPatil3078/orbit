package cmd

import (
	"os"
	"testing"
)

func TestRunCmd_Success(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "orbit-test-run")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Create a valid orbit.yaml
	yamlContent := []byte(`tasks:
  test_task:
    command: "echo test"
    depends_on: []
`)
	os.WriteFile("orbit.yaml", yamlContent, 0644)

	// Run the command
	runCmd.Run(runCmd, []string{})
}
