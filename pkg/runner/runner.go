package runner

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"sync"

	"orbit/pkg/graph"
	"github.com/bmatcuk/doublestar/v4"
)

type TaskStatus string

const (
	StatusSuccess TaskStatus = "Success"
	StatusFailed  TaskStatus = "Failed"
	StatusSkipped TaskStatus = "Skipped"
)

type TaskResult struct {
	Node   *graph.Node
	Output string
	Error  error
	Status TaskStatus
}

func ExecuteTiers(tiers [][]*graph.Node, changedFiles []string, forceAll bool) error {
	skipStates := make(map[string]bool)

	for i, tier := range tiers {
		fmt.Printf("Executing Tier %d (%d tasks)...\n", i, len(tier))

		var wg sync.WaitGroup
		results := make(chan TaskResult, len(tier))

		for _, node := range tier {
			wg.Add(1)
			go func(n *graph.Node) {
				defer wg.Done()
				
				shouldSkip := false
				if !forceAll {
					hasOwnFilters := len(n.TriggerPaths) > 0 || len(n.IgnorePaths) > 0
					
					if hasOwnFilters {
						shouldSkip = true
						for _, file := range changedFiles {
							ignored := false
							for _, ignoreGlob := range n.IgnorePaths {
								match, _ := doublestar.Match(ignoreGlob, file)
								if match {
									ignored = true
									break
								}
							}
							if ignored {
								continue
							}
							
							if len(n.TriggerPaths) == 0 {
								shouldSkip = false
								break
							}
							
							for _, triggerGlob := range n.TriggerPaths {
								match, _ := doublestar.Match(triggerGlob, file)
								if match {
									shouldSkip = false
									break
								}
							}
							if !shouldSkip {
								break
							}
						}
					} else {
						for _, dep := range n.DependsOn {
							if skipStates[dep] {
								shouldSkip = true
								break
							}
						}
					}
				}

				if shouldSkip {
					results <- TaskResult{Node: n, Status: StatusSkipped}
					return
				}

				var cmd *exec.Cmd
				if runtime.GOOS == "windows" {
					cmd = exec.Command("cmd", "/C", n.Command)
				} else {
					cmd = exec.Command("sh", "-c", n.Command)
				}
				if n.WorkingDir != "" {
					cmd.Dir = n.WorkingDir
				}
				
				var outBuf bytes.Buffer
				cmd.Stdout = &outBuf
				cmd.Stderr = &outBuf

				err := cmd.Run()
				status := StatusSuccess
				if err != nil {
					status = StatusFailed
				}
				results <- TaskResult{Node: n, Output: outBuf.String(), Error: err, Status: status}
			}(node)
		}

		wg.Wait()
		close(results)

		var tierErrors []error
		for res := range results {
			if res.Status == StatusSkipped {
				skipStates[res.Node.Name] = true
				fmt.Printf("[-] Task '%s' skipped (No relevant files changed).\n", res.Node.Name)
			} else if res.Status == StatusFailed {
				skipStates[res.Node.Name] = false
				fmt.Printf("\n[X] Task '%s' failed:\n%s\n", res.Node.Name, res.Output)
				tierErrors = append(tierErrors, fmt.Errorf("task %s failed", res.Node.Name))
			} else {
				skipStates[res.Node.Name] = false
				fmt.Printf("[V] Task '%s' finished successfully.\n", res.Node.Name)
			}
		}

		if len(tierErrors) > 0 {
			if len(tierErrors) == 1 {
				return tierErrors[0]
			}
			return fmt.Errorf("%d tasks failed in tier %d", len(tierErrors), i)
		}
	}
	return nil
}
