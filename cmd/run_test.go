package cmd

import (
	"os"
	"os/exec"
	"strings"
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
    command: "echo test > output.txt"
    depends_on: []
`)
	os.WriteFile("orbit.yaml", yamlContent, 0644)
	exec.Command("git", "init").Run()

	// Run the command
	runCmd.Run(runCmd, []string{})

	// Verify the command actually ran by checking the output file
	content, err := os.ReadFile("output.txt")
	if err != nil {
		t.Fatalf("Expected output.txt to be created by the task, got error: %v", err)
	}
	
	if strings.TrimSpace(string(content)) != "test" {
		t.Errorf("Expected output.txt to contain 'test', got %q", string(content))
	}
}

