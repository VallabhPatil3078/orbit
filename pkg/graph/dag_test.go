package graph

import (
	"testing"
)

func TestTopologicalSort_Success(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("lint", "npm run lint", []string{})
	dag.AddNode("format", "prettier", []string{})
	dag.AddNode("build", "npm run build", []string{"lint", "format"})
	dag.AddNode("test", "npm run test", []string{"build"})

	err := dag.BuildEdges()
	if err != nil {
		t.Fatalf("unexpected error building edges: %v", err)
	}

	tiers, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected error sorting DAG: %v", err)
	}

	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}

	// Tier 0 should have lint and format (order doesn't matter, but both must be there)
	if len(tiers[0]) != 2 {
		t.Errorf("expected 2 tasks in tier 0, got %d", len(tiers[0]))
	}

	// Tier 1 should have build
	if len(tiers[1]) != 1 || tiers[1][0].Name != "build" {
		t.Errorf("expected build in tier 1")
	}

	// Tier 2 should have test
	if len(tiers[2]) != 1 || tiers[2][0].Name != "test" {
		t.Errorf("expected test in tier 2")
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("taskA", "echo A", []string{"taskB"})
	dag.AddNode("taskB", "echo B", []string{"taskA"})

	err := dag.BuildEdges()
	if err != nil {
		t.Fatalf("unexpected error building edges: %v", err)
	}

	_, err = dag.TopologicalSort()
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}
