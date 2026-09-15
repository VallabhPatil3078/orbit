package runner

import (
	"orbit/pkg/graph"
	"testing"
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
