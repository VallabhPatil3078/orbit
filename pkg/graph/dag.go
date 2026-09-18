package graph

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// Node represents a task in the DAG.
type Node struct {
	Name         string
	Command      string
	WorkingDir   string
	DependsOn    []string
	TriggerPaths []string
	IgnorePaths  []string
	InDegree     int
	Timeout      time.Duration
}

// DAG represents the Directed Acyclic Graph.
type DAG struct {
	Nodes     map[string]*Node
	Adjacency map[string][]string // Maps a node to the nodes that depend on it
}

// NewDAG creates an empty DAG.
func NewDAG() *DAG {
	return &DAG{
		Nodes:     make(map[string]*Node),
		Adjacency: make(map[string][]string),
	}
}

func (g *DAG) AddNode(name, command, workingDir string, dependsOn, triggerPaths, ignorePaths []string, timeout string) error {
	var d time.Duration
	var err error
	if timeout != "" {
		d, err = time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("task %q has invalid timeout %q: %v", name, timeout, err)
		}
		if d <= 0 {
			return fmt.Errorf("task %q has invalid timeout %q: must be positive", name, timeout)
		}
		if d > 1*time.Hour {
			// Just a warning
			fmt.Printf("[WARNING] Task %q has an unusually long timeout: %v\n", name, d)
		}
	}

	g.Nodes[name] = &Node{
		Name:         name,
		Command:      command,
		WorkingDir:   workingDir,
		DependsOn:    dependsOn,
		TriggerPaths: triggerPaths,
		IgnorePaths:  ignorePaths,
		InDegree:     0,
		Timeout:      d,
	}
	return nil
}

// BuildEdges calculates the in-degrees and builds the adjacency list.
func (g *DAG) BuildEdges() error {
	for name, node := range g.Nodes {
		for _, dep := range node.DependsOn {
			if _, exists := g.Nodes[dep]; !exists {
				return fmt.Errorf("task %q depends on unknown task %q", name, dep)
			}
			g.Adjacency[dep] = append(g.Adjacency[dep], name)
			node.InDegree++
		}
	}
	return nil
}

// TopologicalSort performs Kahn's Algorithm to sort the DAG into execution tiers.
func (g *DAG) TopologicalSort() ([][]*Node, error) {
	var tiers [][]*Node
	
	// Find all nodes with in-degree 0 (no dependencies)
	var currentTier []*Node
	for _, node := range g.Nodes {
		if node.InDegree == 0 {
			currentTier = append(currentTier, node)
		}
	}

	processedCount := 0

	for len(currentTier) > 0 {
		sort.Slice(currentTier, func(i, j int) bool {
			return currentTier[i].Name < currentTier[j].Name
		})

		tiers = append(tiers, currentTier)
		processedCount += len(currentTier)

		var nextTier []*Node
		for _, node := range currentTier {
			// For every node that depends on the current node
			for _, dependentName := range g.Adjacency[node.Name] {
				dependentNode := g.Nodes[dependentName]
				dependentNode.InDegree--
				if dependentNode.InDegree == 0 {
					nextTier = append(nextTier, dependentNode)
				}
			}
		}
		currentTier = nextTier
	}

	if processedCount != len(g.Nodes) {
		return nil, errors.New("cycle detected in DAG")
	}

	return tiers, nil
}

