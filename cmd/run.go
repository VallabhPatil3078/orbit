package cmd

import (
	"fmt"
	"os"
	
	"orbit/pkg/config"
	"orbit/pkg/graph"
	"orbit/pkg/runner"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes the DAG pipeline",
	Long:  `Parses orbit.yaml, builds the DAG, and executes the tasks concurrently.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("-> Starting Orbit Pipeline...")

		// 1. Parse config
		cfg, err := config.ParseConfig("orbit.yaml")
		if err != nil {
			fmt.Printf("[X] Failed to load orbit.yaml: %v\n", err)
			os.Exit(1)
		}

		// 2. Build DAG
		dag := graph.NewDAG()
		for name, taskCfg := range cfg.Tasks {
			dag.AddNode(name, taskCfg.Command, taskCfg.WorkingDir, taskCfg.DependsOn, taskCfg.TriggerPaths, taskCfg.IgnorePaths)
		}

		if err := dag.BuildEdges(); err != nil {
			fmt.Printf("[X] Invalid dependencies: %v\n", err)
			os.Exit(1)
		}

		// 3. Topological Sort
		tiers, err := dag.TopologicalSort()
		if err != nil {
			fmt.Printf("[X] DAG Error: %v\n", err)
			os.Exit(1)
		}

		// 4. Execute Tiers
		if err := runner.ExecuteTiers(tiers); err != nil {
			fmt.Printf("\n[X] Orbit pipeline failed!\n")
			os.Exit(1)
		}

		fmt.Println("\n-> Orbit pipeline completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

