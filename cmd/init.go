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
		fmt.Println("Orbit is ready to go! Edit orbit.yaml and try committing.")
	},
}

func init() {
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing orbit.yaml and hooks")
	rootCmd.AddCommand(initCmd)
}
