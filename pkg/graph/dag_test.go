package graph

import (
	"strings"
	"testing"
)

func TestTopologicalSort_Success(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("lint", "npm run lint", "", []string{}, nil, nil, "")
	dag.AddNode("format", "prettier", "", []string{}, nil, nil, "")
	dag.AddNode("build", "npm run build", "", []string{"lint", "format"}, nil, nil, "")
	dag.AddNode("test", "npm run test", "", []string{"build"}, nil, nil, "")

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

	if len(tiers[0]) != 2 {
		t.Errorf("expected 2 tasks in tier 0, got %d", len(tiers[0]))
	}

	if len(tiers[1]) != 1 || tiers[1][0].Name != "build" {
		t.Errorf("expected build in tier 1")
	}

	if len(tiers[2]) != 1 || tiers[2][0].Name != "test" {
		t.Errorf("expected test in tier 2")
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("taskA", "echo A", "", []string{"taskB"}, nil, nil, "")
	dag.AddNode("taskB", "echo B", "", []string{"taskA"}, nil, nil, "")

	err := dag.BuildEdges()
	if err != nil {
		t.Fatalf("unexpected error building edges: %v", err)
	}

	_, err = dag.TopologicalSort()
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestBuildEdges_MissingDependency(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("build", "npm run build", "", []string{"lnit"}, nil, nil, "") // intentional typo

	err := dag.BuildEdges()
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
	expectedErrMsg := "task \"build\" depends on unknown task \"lnit\""
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("expected error message to contain %q, got %q", expectedErrMsg, err.Error())
	}
}

func TestTopologicalSort_Diamond(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("A", "echo A", "", []string{}, nil, nil, "")
	dag.AddNode("B", "echo B", "", []string{"A"}, nil, nil, "")
	dag.AddNode("C", "echo C", "", []string{"A"}, nil, nil, "")
	dag.AddNode("D", "echo D", "", []string{"B", "C"}, nil, nil, "")

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

	// Tier 0
	if len(tiers[0]) != 1 || tiers[0][0].Name != "A" {
		t.Errorf("expected tier 0 to have [A]")
	}

	// Tier 1
	if len(tiers[1]) != 2 {
		t.Errorf("expected tier 1 to have [B, C]")
	}

	// Tier 2
	if len(tiers[2]) != 1 || tiers[2][0].Name != "D" {
		t.Errorf("expected tier 2 to have [D]")
	}
}

