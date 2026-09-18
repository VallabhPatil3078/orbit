package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "orbit",
	Short: "Orbit is a local DAG-based task orchestrator for Git hooks",
	Long: `Orbit intelligently schedules and executes dependent validation tasks 
concurrently to maximize CPU efficiency during your Git pre-commit workflow.`,
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
