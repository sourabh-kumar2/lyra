package graph

import (
	"fmt"
	"testing"
)

func BenchmarkDependencyDAG(b *testing.B) {
	benchmarks := []struct {
		name string
		deps map[string][]string
	}{
		{
			name: "small graph 10 nodes",
			deps: generateLinearGraph(10),
		},
		{
			name: "medium graph 100 nodes",
			deps: generateLinearGraph(100),
		},
		{
			name: "large graph 1000 nodes",
			deps: generateLinearGraph(1000),
		},
		{
			name: "wide graph 100 nodes parallel",
			deps: generateWideGraph(100),
		},
		{
			name: "deep graph 100 levels",
			deps: generateDeepGraph(100),
		},
		{
			name: "complex diamond 50 nodes",
			deps: generateDiamondGraph(50),
		},
		{
			name: "complex diamond 100 nodes",
			deps: generateComplexGraph(200),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			dag := NewDependencyDAG(bm.deps)

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_, err := dag.GetExecutionLevels()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Graph generators for different topologies.
func generateLinearGraph(size int) map[string][]string {
	deps := make(map[string][]string)

	for i := range size {
		nodeID := nodeIDFromInt(i)
		if i == 0 {
			deps[nodeID] = []string{}
		} else {
			deps[nodeID] = []string{nodeIDFromInt(i - 1)}
		}
	}

	return deps
}

func generateWideGraph(size int) map[string][]string {
	deps := make(map[string][]string)

	// Root node
	deps["root"] = []string{}

	// All other nodes depend on root
	for i := 1; i < size; i++ {
		deps[nodeIDFromInt(i)] = []string{"root"}
	}

	return deps
}

func generateDeepGraph(depth int) map[string][]string {
	deps := make(map[string][]string)

	for i := range depth {
		nodeID := nodeIDFromInt(i)
		if i == 0 {
			deps[nodeID] = []string{}
		} else {
			deps[nodeID] = []string{nodeIDFromInt(i - 1)}
		}
	}

	return deps
}

//revive:disable-next-line:cognitive-complexity
func generateDiamondGraph(layers int) map[string][]string {
	deps := make(map[string][]string)

	for layer := range layers {
		switch layer {
		case 0:
			// Root
			deps[nodeIDFromInt(0)] = []string{}
		case layers - 1:
			// Final node depends on all nodes in previous layer
			var prevNodes []string
			start := (layer-1)*2 + 1
			for j := range 2 {
				prevNodes = append(prevNodes, nodeIDFromInt(start+j))
			}
			deps[nodeIDFromInt(layer*2+1)] = prevNodes
		default:
			// Each layer has 2 nodes depending on previous layer
			prevLayerStart := (layer-1)*2 + 1
			if layer == 1 {
				prevLayerStart = 0
			}

			for j := range 2 {
				nodeID := nodeIDFromInt(layer*2 + j + 1)
				if layer == 1 {
					deps[nodeID] = []string{nodeIDFromInt(0)}
				} else {
					deps[nodeID] = []string{nodeIDFromInt(prevLayerStart), nodeIDFromInt(prevLayerStart + 1)}
				}
			}
		}
	}

	return deps
}

func generateComplexGraph(size int) map[string][]string {
	deps := make(map[string][]string)

	// Mix of patterns: some linear chains, some diamonds, some wide dependencies
	for i := range size {
		nodeID := nodeIDFromInt(i)

		switch {
		case i == 0:
			deps[nodeID] = []string{}
		case i < size/4: // Linear chain
			deps[nodeID] = []string{nodeIDFromInt(i - 1)}
		case i < size/2: // Wide dependencies
			deps[nodeID] = []string{nodeIDFromInt(0), nodeIDFromInt(1)}
		case i < 3*size/4: // Diamond pattern
			deps[nodeID] = []string{nodeIDFromInt(i - 2), nodeIDFromInt(i - 3)}
		default: // Complex dependencies
			var depList []string
			for j := 0; j < 3 && i-j-1 >= 0; j++ {
				depList = append(depList, nodeIDFromInt(i-j-1))
			}
			deps[nodeID] = depList
		}
	}

	return deps
}

func nodeIDFromInt(i int) string {
	return fmt.Sprintf("node_%d", i)
}

// Memory allocation benchmarks.
func BenchmarkDependencyDAGMemory(b *testing.B) {
	deps := generateComplexGraph(100)

	b.Run("memory allocations", func(b *testing.B) {
		dag := NewDependencyDAG(deps)

		b.ReportAllocs()
		b.ResetTimer()

		for range b.N {
			_, err := dag.GetExecutionLevels()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
