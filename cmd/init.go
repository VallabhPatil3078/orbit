package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes Orbit in the current repository",
	Long:  `Generates the orbit.yaml config file and installs the Git pre-commit hook.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Create orbit.yaml template
		yamlContent := []byte(`tasks:
  hello:
    command: "echo Hello World"
    depends_on: []
`)
		if err := os.WriteFile("orbit.yaml", yamlContent, 0644); err != nil {
			fmt.Printf("Error creating orbit.yaml: %v\n", err)
			return
		}
		fmt.Println("[V] Created orbit.yaml")

		// 2. Create the pre-commit hook
		hookDir := filepath.Join(".git", "hooks")
		if _, err := os.Stat(hookDir); os.IsNotExist(err) {
			fmt.Println("[!] Not a git repository or .git/hooks missing. Skipping hook installation.")
			return
		}

		hookPath := filepath.Join(hookDir, "pre-commit")
		
		// For Windows, it's typically a shell script that Git Bash runs
		hookContent := []byte("#!/bin/sh\nexec orbit run\n")
		
		if err := os.WriteFile(hookPath, hookContent, 0755); err != nil {
			fmt.Printf("Error creating pre-commit hook: %v\n", err)
			return
		}
		
		fmt.Println("[V] Successfully installed Git pre-commit hook!")
		fmt.Println("Orbit is ready to go! Edit orbit.yaml and try committing.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
