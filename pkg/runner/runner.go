package runner

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"orbit/pkg/graph"
)

// TaskResult holds the outcome of an executed task.
type TaskResult struct {
	Node   *graph.Node
	Output string
	Error  error
}

// ExecuteTiers takes the topologically sorted tiers of the DAG and runs them.
func ExecuteTiers(tiers [][]*graph.Node) error {
	for i, tier := range tiers {
		fmt.Printf("Executing Tier %d (%d tasks)...\n", i, len(tier))

		var wg sync.WaitGroup
		results := make(chan TaskResult, len(tier))

		// Launch a Goroutine for each task in the current tier
		for _, node := range tier {
			wg.Add(1)
			go runTask(node, &wg, results)
		}

		// Wait for all tasks in this tier to finish before moving to the next tier
		wg.Wait()
		close(results)

		var tierErrors []error
		// Check results for this tier
		for res := range results {
			if res.Error != nil {
				// Print the buffered output if it failed so the user knows what went wrong
				fmt.Printf("\n[X] Task '%s' failed:\n%s\n", res.Node.Name, res.Output)
				tierErrors = append(tierErrors, fmt.Errorf("task %s failed", res.Node.Name))
			} else {
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

func runTask(node *graph.Node, wg *sync.WaitGroup, results chan<- TaskResult) {
	defer wg.Done()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", node.Command)
	} else {
		cmd = exec.Command("sh", "-c", node.Command)
	}
	
	// We capture stdout and stderr together to buffer it
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	results <- TaskResult{
		Node:   node,
		Output: outBuf.String(),
		Error:  err,
	}
}
