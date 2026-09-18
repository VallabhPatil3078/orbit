package runner

import (
	"context"
	"testing"
	"time"

	"github.com/VallabhPatil3078/orbit/pkg/graph"
)

func TestExecuteTiers_SkipLogic(t *testing.T) {
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
			name:         "No changed files (A1 Fix: runs all as fallback)",
			changedFiles: []string{},
			forceAll:     false,
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
			rep := NewMockReporter()
			err := ExecuteTiers(context.Background(), tiers, tc.changedFiles, tc.forceAll, "", rep, 10*time.Minute, 5*time.Second)
			if err != nil {
				t.Fatalf("ExecuteTiers failed: %v", err)
			}

			for task, shouldSkip := range tc.expectedSkip {
				hasSkip := rep.SkippedTasks[task]
				hasRun := rep.SucceededTasks[task]
				
				if shouldSkip {
					if !hasSkip {
						t.Errorf("Expected task %s to skip, but it did not.", task)
					}
					if hasRun {
						t.Errorf("Expected task %s to skip, but it ran.", task)
					}
				} else {
					if !hasRun {
						t.Errorf("Expected task %s to run, but it did not.", task)
					}
					if hasSkip {
						t.Errorf("Expected task %s to run, but it skipped.", task)
					}
				}
			}
		})
	}
}

func TestExecuteTiers_MultiParentSkip(t *testing.T) {
	nodeX := &graph.Node{Name: "X", Command: "echo X", TriggerPaths: []string{"**/*.go"}}
	nodeY := &graph.Node{Name: "Y", Command: "echo Y", TriggerPaths: []string{"**/*.md"}}
	nodeChild := &graph.Node{Name: "Child", Command: "echo Child", DependsOn: []string{"X", "Y"}}
	
	tiers := [][]*graph.Node{{nodeX, nodeY}, {nodeChild}}

	t.Run("Only X runs", func(t *testing.T) {
		rep := NewMockReporter()
		ExecuteTiers(context.Background(), tiers, []string{"main.go"}, false, "", rep, 10*time.Minute, 5*time.Second)

		if rep.SkippedTasks["Child"] {
			t.Errorf("Child should not skip when one parent (X) runs")
		}
	})
	
	t.Run("Both X and Y skip", func(t *testing.T) {
		rep := NewMockReporter()
		ExecuteTiers(context.Background(), tiers, []string{"unrelated.txt"}, false, "", rep, 10*time.Minute, 5*time.Second)

		if !rep.SkippedTasks["Child"] {
			t.Errorf("Child should skip when all parents skip")
		}
	})
}

func TestExecuteTiers_BothFilters(t *testing.T) {
	nodeA := &graph.Node{Name: "A", Command: "echo A", TriggerPaths: []string{"**/*.go"}, IgnorePaths: []string{"vendor/**"}}
	tiers := [][]*graph.Node{{nodeA}}

	t.Run("Trigger but ignored", func(t *testing.T) {
		rep := NewMockReporter()
		ExecuteTiers(context.Background(), tiers, []string{"vendor/main.go"}, false, "", rep, 10*time.Minute, 5*time.Second)

		if !rep.SkippedTasks["A"] {
			t.Errorf("Task should skip if file is ignored")
		}
	})
	
	t.Run("Trigger not ignored", func(t *testing.T) {
		rep := NewMockReporter()
		ExecuteTiers(context.Background(), tiers, []string{"src/main.go"}, false, "", rep, 10*time.Minute, 5*time.Second)

		if rep.SkippedTasks["A"] {
			t.Errorf("Task should run if file is triggered and not ignored")
		}
	})
}
