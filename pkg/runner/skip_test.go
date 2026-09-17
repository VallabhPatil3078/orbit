package runner

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"orbit/pkg/graph"
)

func TestExecuteTiers_SkipLogic(t *testing.T) {
	// Task A: trigger on src/**
	// Task B: depends on A, no own triggers (should cascade skip if A skipped)
	// Task C: depends on A, but has own triggers test/** (evaluates independently)
	
	nodeA := &graph.Node{
		Name:         "A",
		Command:      "echo A",
		TriggerPaths: []string{"src/**"},
	}
	
	nodeB := &graph.Node{
		Name:      "B",
		Command:   "echo B",
		DependsOn: []string{"A"},
	}
	
	nodeC := &graph.Node{
		Name:         "C",
		Command:      "echo C",
		DependsOn:    []string{"A"},
		TriggerPaths: []string{"test/**"},
	}
	
	tiers := [][]*graph.Node{{nodeA}, {nodeB, nodeC}}
	
	tests := []struct {
		name         string
		changedFiles []string
		forceAll     bool
		expectedSkip map[string]bool
	}{
		{
			name:         "Force All (no skips)",
			changedFiles: []string{},
			forceAll:     true,
			expectedSkip: map[string]bool{"A": false, "B": false, "C": false},
		},
		{
			name:         "Change in src (A runs, B runs due to dependency, C skips due to own filter)",
			changedFiles: []string{"src/main.go"},
			forceAll:     false,
			expectedSkip: map[string]bool{"A": false, "B": false, "C": true},
		},
		{
			name:         "Change in test (A skips, B skips due to cascade, C runs due to own filter)",
			changedFiles: []string{"test/main_test.go"},
			forceAll:     false,
			expectedSkip: map[string]bool{"A": true, "B": true, "C": false},
		},
		{
			name:         "Change in unmonitored path (All skip)",
			changedFiles: []string{"docs/readme.md"},
			forceAll:     false,
			expectedSkip: map[string]bool{"A": true, "B": true, "C": true},
		},
		{
			name:         "Change in both src and test (All run)",
			changedFiles: []string{"src/main.go", "test/main_test.go"},
			forceAll:     false,
			expectedSkip: map[string]bool{"A": false, "B": false, "C": false},
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Redirect stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := ExecuteTiers(tiers, tc.changedFiles, tc.forceAll)
			if err != nil {
				t.Fatalf("ExecuteTiers failed: %v", err)
			}

			w.Close()
			os.Stdout = oldStdout
			
			var buf bytes.Buffer
			io.Copy(&buf, r)
			out := buf.String()

			for task, shouldSkip := range tc.expectedSkip {
				skipMsg := "[-] Task '" + task + "' skipped"
				runMsg := "[V] Task '" + task + "' finished"
				
				hasSkip := strings.Contains(out, skipMsg)
				hasRun := strings.Contains(out, runMsg)
				
				if shouldSkip {
					if !hasSkip {
						t.Errorf("Expected task %s to skip, but it did not.\nOutput:\n%s", task, out)
					}
					if hasRun {
						t.Errorf("Expected task %s to skip, but it ran.\nOutput:\n%s", task, out)
					}
				} else {
					if !hasRun {
						t.Errorf("Expected task %s to run, but it did not.\nOutput:\n%s", task, out)
					}
					if hasSkip {
						t.Errorf("Expected task %s to run, but it skipped.\nOutput:\n%s", task, out)
					}
				}
			}
		})
	}
}
