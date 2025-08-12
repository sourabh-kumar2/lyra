package graph

import (
	"github.com/sourabh-kumar2/lyra/errors"
)

// DependencyDAG represents a directed acyclic graph for managing task dependencies.
// It uses Kahn's algorithm for topological sorting and cycle detection.
type DependencyDAG struct {
	deps map[string][]string
}

// NewDependencyDAG creates a dependency dag object.
func NewDependencyDAG(dependencies map[string][]string) *DependencyDAG {
	return &DependencyDAG{
		deps: dependencies,
	}
}

// GetExecutionLevels returns the nodes grouped by execution levels using Kahn's algorithm.
// Nodes in the same level can be executed concurrently as they have no dependencies between them.
// Returns an error if cycles are detected or if missing dependencies are found.
//
//nolint:cyclop // Kahn's algo
//revive:disable-next-line:cyclomatic,cognitive-complexity
func (g *DependencyDAG) GetExecutionLevels() ([][]string, error) {
	if len(g.deps) == 0 {
		return [][]string{}, nil
	}

	inDegree, err := g.getInDegree()
	if err != nil {
		return nil, err
	}

	queue := make([]string, 0, len(g.deps))
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	reverseDeps := make(map[string][]string, len(g.deps))
	for nodeID, deps := range g.deps {
		for _, dep := range deps {
			reverseDeps[dep] = append(reverseDeps[dep], nodeID)
		}
	}

	levels := make([][]string, 0, len(g.deps))
	processedCount := 0

	for len(queue) > 0 {
		// Sort current level for deterministic output
		currentLevel := make([]string, len(queue))
		copy(currentLevel, queue)

		levels = append(levels, currentLevel)
		processedCount += len(currentLevel)

		// Process all nodes in current level
		nextQueue := queue[:0]
		for _, nodeID := range queue {
			for _, dependentID := range reverseDeps[nodeID] {
				inDegree[dependentID]--
				if inDegree[dependentID] == 0 {
					nextQueue = append(nextQueue, dependentID)
				}
			}
		}

		queue = nextQueue
	}

	if processedCount != len(g.deps) {
		return nil, errors.ErrCyclicDependency
	}

	return levels, nil
}

func (g *DependencyDAG) getInDegree() (map[string]int, error) {
	inDegree := make(map[string]int)

	// Initialize in-degrees for all nodes
	for nodeID := range g.deps {
		inDegree[nodeID] = 0
	}

	for nodeID, deps := range g.deps {
		for _, depNode := range deps {
			if _, exists := inDegree[depNode]; !exists {
				return nil, errors.Wrapf(
					errors.ErrMissingDependency,
					"node %q depends on non-existent node %q",
					nodeID,
					depNode,
				)
			}
			inDegree[nodeID]++ // nodeID has an incoming edge
		}
	}
	return inDegree, nil
}
