package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var force bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes Orbit in the current repository",
	Long:  `Generates the orbit.yaml config file and installs the Git pre-commit hook.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Create orbit.yaml template
		if _, err := os.Stat("orbit.yaml"); err == nil && !force {
			fmt.Println("[!] orbit.yaml already exists. Skipping. (use --force to overwrite)")
			return
		}

		yamlContent := []byte(`# Orbit configuration file — full docs: https://github.com/VallabhPatil3078/orbit#readme
version: 1

tasks:
  # Tasks with no dependencies run first, in parallel.
  # Replace 'echo' with your real lint/test/build commands.
  lint:
    command: "echo Linting code..."
    depends_on: []
    # Only runs when a matching file has changed (supports ** globs).
    # Adjust to your stack, e.g. "**/*.go", "**/*.py", "src/**/*.ts"
    trigger_paths: ["src/**"]

  test:
    command: "echo Running tests..."
    # Waits for lint. Runs independently since it has its own
    # trigger_paths — see docs on cascade-skip for tasks with none.
    depends_on: ["lint"]
    working_dir: "."          # run from a subdirectory if needed, e.g. "backend"
    timeout: "2m"               # optional — overrides the global default
    trigger_paths: ["src/**", "test/**"]
    ignore_paths: ["**/*.md"]   # excluded from triggering, checked first

# Tip: run "orbit validate" any time to check this file without
# actually executing any tasks.
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
		hookContent := []byte("#!/bin/sh\nexec orbit run\n")

		if _, err := os.Stat(hookPath); err == nil {
			existingContent, readErr := os.ReadFile(hookPath)
			if readErr != nil {
				fmt.Printf("Error reading existing pre-commit hook: %v\n", readErr)
				return
			}
			
			if bytes.Contains(existingContent, []byte("exec orbit run")) {
				fmt.Println("[V] Orbit is already installed in pre-commit hook.")
			} else if !force {
				backupPath := hookPath + ".orbit-backup"
				fmt.Printf("[!] Existing pre-commit hook found. Backing up to %s\n", backupPath)
				if renameErr := os.Rename(hookPath, backupPath); renameErr != nil {
					fmt.Printf("Error backing up hook: %v\n", renameErr)
					return
				}
			}
		}
		
		if err := os.WriteFile(hookPath, hookContent, 0755); err != nil {
			fmt.Printf("Error creating pre-commit hook: %v\n", err)
			return
		}
		
		fmt.Println("[V] Successfully installed Git pre-commit hook!")
		fmt.Println("Orbit is ready! See the comments in orbit.yaml for a full feature walkthrough, or run 'orbit validate' to check your config.")
	},
}

func init() {
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing orbit.yaml and hooks")
	rootCmd.AddCommand(initCmd)
}
