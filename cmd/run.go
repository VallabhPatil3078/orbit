package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"orbit/pkg/config"
	"orbit/pkg/git"
	"orbit/pkg/graph"
	"orbit/pkg/runner"

	"github.com/spf13/cobra"
)

var (
	forceAll bool
	baseRef  string
	quiet    bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes the DAG pipeline",
	Long: `Parses orbit.yaml, builds the DAG, and executes the tasks concurrently.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		var rep runner.Reporter = &runner.TextReporter{}
		if quiet {
			rep = &runner.QuietReporter{}
		}
		rep.PipelineStarted()

		// 0. Get changed files
		changedFiles, err := git.GetChangedFiles(baseRef)
		if err != nil && !forceAll {
			fmt.Printf("[!] Could not get changed files (run with --all to force): %v\n", err)
			os.Exit(1)
		}

		// 1. Parse config
		cfg, err := config.ParseConfig("orbit.yaml")
		if err != nil {
			fmt.Printf("[X] Failed to load orbit.yaml: %v\n", err)
			os.Exit(1) // Exit 1 for Config Error
		}

		// 2. Build DAG
		dag := graph.NewDAG()
		for name, taskCfg := range cfg.Tasks {
			dag.AddNode(name, taskCfg.Command, taskCfg.WorkingDir, taskCfg.DependsOn, taskCfg.TriggerPaths, taskCfg.IgnorePaths)
		}

		if err := dag.BuildEdges(); err != nil {
			fmt.Printf("[X] Invalid dependencies: %v\n", err)
			os.Exit(2) // Exit 2 for DAG Error
		}

		// 3. Topological Sort
		tiers, err := dag.TopologicalSort()
		if err != nil {
			fmt.Printf("[X] DAG Error: %v\n", err)
			os.Exit(2) // Exit 2 for DAG Error
		}

		// 4. Execute Tiers
		if err := runner.ExecuteTiers(ctx, tiers, changedFiles, forceAll, baseRef, rep); err != nil {
			rep.PipelineFinished(false)
			if err == runner.ErrTaskTimeout {
				os.Exit(4) // Exit 4 for Timeout
			}
			os.Exit(3) // Exit 3 for Task Failure
		}

		rep.PipelineFinished(true)
	},
}

func init() {
	runCmd.Flags().BoolVar(&forceAll, "all", false, "Force run all tasks regardless of path filters")
	runCmd.Flags().StringVar(&baseRef, "base", "", "Base git ref to compare against (e.g. main) for path filters")
	runCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential task output")
	rootCmd.AddCommand(runCmd)
}

