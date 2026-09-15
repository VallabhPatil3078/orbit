package graph

import (
	"errors"
)

// Node represents a task in the DAG.
type Node struct {
	Name      string
	Command   string
	DependsOn []string
	InDegree  int
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

// AddNode adds a task to the DAG.
func (g *DAG) AddNode(name, command string, dependsOn []string) {
	g.Nodes[name] = &Node{
		Name:      name,
		Command:   command,
		DependsOn: dependsOn,
		InDegree:  0,
	}
}

// BuildEdges calculates the in-degrees and builds the adjacency list.
func (g *DAG) BuildEdges() error {
	for name, node := range g.Nodes {
		for _, dep := range node.DependsOn {
			if _, exists := g.Nodes[dep]; !exists {
				return errors.New("dependency not found: " + dep)
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
