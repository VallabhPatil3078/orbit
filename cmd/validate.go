package cmd

import (
	"fmt"
	"os"

	"orbit/pkg/config"
	"orbit/pkg/graph"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the orbit.yaml configuration without running tasks",
	Long:  `Parses the orbit.yaml file, validates task definitions, and ensures the dependency graph (DAG) has no cycles or missing dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.ParseConfig("orbit.yaml")
		if err != nil {
			fmt.Printf("[X] Config validation failed:\n%v\n", err)
			os.Exit(1)
		}

		dag := graph.NewDAG()
		for name, task := range cfg.Tasks {
			for _, pat := range task.TriggerPaths {
				if _, err := doublestar.Match(pat, "test"); err != nil {
					fmt.Printf("[X] Task %q has invalid trigger path pattern %q: %v\n", name, pat, err)
					os.Exit(1)
				}
			}
			for _, pat := range task.IgnorePaths {
				if _, err := doublestar.Match(pat, "test"); err != nil {
					fmt.Printf("[X] Task %q has invalid ignore path pattern %q: %v\n", name, pat, err)
					os.Exit(1)
				}
			}
			dag.AddNode(name, task.Command, task.WorkingDir, task.DependsOn, task.TriggerPaths, task.IgnorePaths)
		}

		if err := dag.BuildEdges(); err != nil {
			fmt.Printf("[X] DAG validation failed:\n%v\n", err)
			os.Exit(1)
		}

		_, err = dag.TopologicalSort()
		if err != nil {
			fmt.Printf("[X] DAG topological sort failed:\n%v\n", err)
			os.Exit(1)
		}

		fmt.Println("[V] orbit.yaml is perfectly valid! No cycles or missing dependencies detected.")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}



