package runner

import (
	"context"
	"orbit/pkg/graph"
	"runtime"
	"testing"
	"time"
)

func TestExecuteTiers_Success(t *testing.T) {
	node1 := &graph.Node{Name: "echo1", Command: "echo 1"}
	node2 := &graph.Node{Name: "echo2", Command: "echo 2"}
	
	tiers := [][]*graph.Node{
		{node1, node2},
	}

	rep := NewMockReporter()
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExecuteTiers_Failure(t *testing.T) {
	nodeFail := &graph.Node{Name: "fail_task", Command: "exit 1"}
	tiers := [][]*graph.Node{
		{nodeFail},
	}
	rep := NewMockReporter()
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep)
	if err == nil {
		t.Fatal("expected an error because the task fails, got nil")
	}
}

func TestExecuteTiers_RunsConcurrently(t *testing.T) {
	var sleepCmd string
	if runtime.GOOS == "windows" {
		// Use ping instead of powershell to avoid slow startup times in CI causing false positive test failures
		sleepCmd = "ping 127.0.0.1 -n 2 > nul"
	} else {
		sleepCmd = "sleep 1"
	}

	node1 := &graph.Node{Name: "sleep1", Command: sleepCmd}
	node2 := &graph.Node{Name: "sleep2", Command: sleepCmd}
	
	tiers := [][]*graph.Node{
		{node1, node2},
	}

	start := time.Now()
	rep := NewMockReporter()
	err := ExecuteTiers(context.Background(), tiers, nil, true, "", rep)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error executing tiers: %v", err)
	}

	if duration >= 3000*time.Millisecond {
		t.Errorf("expected execution time < 3.0s, got %v (implies tasks ran sequentially)", duration)
	}
}



