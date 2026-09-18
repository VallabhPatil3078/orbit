package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/VallabhPatil3078/orbit/pkg/graph"
	"github.com/bmatcuk/doublestar/v4"
)

type TaskStatus string

const (
	StatusSuccess TaskStatus = "Success"
	StatusFailed  TaskStatus = "Failed"
	StatusSkipped TaskStatus = "Skipped"
)

var ErrTaskTimeout = fmt.Errorf("one or more tasks timed out")

type TaskResult struct {
	Node   *graph.Node
	Output string
	Error  error
	CtxErr error
	Status TaskStatus
}

func ExecuteTiers(ctx context.Context, tiers [][]*graph.Node, changedFiles []string, forceAll bool, baseRef string, rep Reporter, taskTimeout time.Duration, taskGracePeriod time.Duration) error {
	skipStates := make(map[string]bool)

	if len(changedFiles) == 0 && baseRef == "" && !forceAll {
		fmt.Println("[!] No staged changes detected — running all tasks as a fallback.")
		forceAll = true
	}

	for i, tier := range tiers {
		if rep != nil {
			var taskNames []string
			for _, node := range tier {
				taskNames = append(taskNames, node.Name)
			}
			rep.TierStarted(i, taskNames)
		}

		var wg sync.WaitGroup
		results := make(chan TaskResult, len(tier))

		for _, node := range tier {
			if rep != nil {
				rep.TaskStarted(node.Name)
			}
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
								match, err := doublestar.Match(ignoreGlob, file)
								if err != nil {
									fmt.Printf("[!] Warning: invalid ignore path pattern %q: %v\n", ignoreGlob, err)
								}
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
								match, err := doublestar.Match(triggerGlob, file)
								if err != nil {
									fmt.Printf("[!] Warning: invalid trigger path pattern %q: %v\n", triggerGlob, err)
								}
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
						if len(n.DependsOn) > 0 {
							allSkipped := true
							for _, dep := range n.DependsOn {
								if !skipStates[dep] {
									allSkipped = false
									break
								}
							}
							shouldSkip = allSkipped
						}
					}
				}

				if shouldSkip {
					results <- TaskResult{Node: n, Status: StatusSkipped}
					return
				}

				timeout := taskTimeout
				if n.Timeout > 0 {
					timeout = n.Timeout
				}
				taskCtx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

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

				pt := NewProcessTree(cmd)
				err := pt.Start()
				
				var waitErr error
				if err == nil {
					done := make(chan struct{})
					go func() {
						select {
						case <-taskCtx.Done():
							pt.Kill(taskGracePeriod)
						case <-done:
						}
					}()
					
					waitErr = pt.Wait()
					close(done)
				} else {
					waitErr = err
				}

				status, ctxErr := DetermineTaskStatus(waitErr, taskCtx.Err())
				results <- TaskResult{Node: n, Output: outBuf.String(), Error: waitErr, CtxErr: ctxErr, Status: status}
			}(node)
		}

		wg.Wait()
		close(results)

		resMap := make(map[string]TaskResult)
		for res := range results {
			resMap[res.Node.Name] = res
		}

		var tierErrors []error
		hasTimeout := false
		for _, node := range tier {
			res := resMap[node.Name]
			if res.Status == StatusSkipped {
				skipStates[res.Node.Name] = true
				if rep != nil {
					rep.TaskSkipped(res.Node.Name, "no relevant files changed")
				}
			} else if res.Status == StatusFailed {
				skipStates[res.Node.Name] = false
				if rep != nil {
					rep.TaskFailed(res.Node.Name, res.Output)
				}
				if res.CtxErr == context.DeadlineExceeded {
					hasTimeout = true
					tierErrors = append(tierErrors, fmt.Errorf("task %s timed out", res.Node.Name))
				} else {
					tierErrors = append(tierErrors, fmt.Errorf("task %s failed", res.Node.Name))
				}
			} else {
				skipStates[res.Node.Name] = false
				if rep != nil {
					rep.TaskSucceeded(res.Node.Name)
				}
			}
		}

		if len(tierErrors) > 0 {
			if hasTimeout {
				return ErrTaskTimeout
			}
			if len(tierErrors) == 1 {
				return tierErrors[0]
			}
			return fmt.Errorf("%d tasks failed in tier %d", len(tierErrors), i)
		}
	}
	return nil
}

// DetermineTaskStatus evaluates the exit error and context error to determine the final task status.
// It explicitly handles the race condition where a process exits normally exactly as the timeout hits.
func DetermineTaskStatus(waitErr error, ctxErr error) (TaskStatus, error) {
	if waitErr == nil {
		// If the process completed successfully, ignore any context errors.
		// This prevents a race condition where the process exits normally exactly as the timeout hits.
		return StatusSuccess, nil
	}
	return StatusFailed, ctxErr
}
