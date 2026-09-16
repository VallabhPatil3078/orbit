package runner

import (
	"orbit/pkg/graph"
	"testing"
	"time"
)

func TestExecuteTiers_Success(t *testing.T) {
	node1 := &graph.Node{Name: "echo1", Command: "echo 1"}
	node2 := &graph.Node{Name: "echo2", Command: "echo 2"}
	
	// Create a dummy tier with two independent tasks
	tiers := [][]*graph.Node{
		{node1, node2},
	}

	err := ExecuteTiers(tiers)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExecuteTiers_Failure(t *testing.T) {
	// this command will fail
	nodeFail := &graph.Node{Name: "fail_task", Command: "exit 1"}
	
	tiers := [][]*graph.Node{
		{nodeFail},
	}

	err := ExecuteTiers(tiers)
	if err == nil {
		t.Fatal("expected an error because the task fails, got nil")
	}
}

func TestExecuteTiers_RunsConcurrently(t *testing.T) {
	// We use PowerShell to sleep for 1 second.
	// If two tasks run sequentially, it will take ~2 seconds.
	// If they run concurrently, it will take ~1 second.
	node1 := &graph.Node{Name: "sleep1", Command: "powershell -c \"Start-Sleep 1\""}
	node2 := &graph.Node{Name: "sleep2", Command: "powershell -c \"Start-Sleep 1\""}
	
	tiers := [][]*graph.Node{
		{node1, node2},
	}

	start := time.Now()
	err := ExecuteTiers(tiers)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error executing tiers: %v", err)
	}

	// We assert that it takes less than 1.5 seconds.
	if duration >= 1500*time.Millisecond {
		t.Errorf("expected execution time < 1.5s, got %v (implies tasks ran sequentially)", duration)
	}
}
