package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes the DAG pipeline",
	Long:  `Parses orbit.yaml, builds the DAG, and executes the tasks concurrently.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Orbit DAG... (Coming soon)")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
