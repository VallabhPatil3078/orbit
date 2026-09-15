package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes Orbit in the current repository",
	Long:  `Generates the orbit.yaml config file and installs the Git pre-commit hook.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Orbit initialized! (Coming soon: creating orbit.yaml and git hook)")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
