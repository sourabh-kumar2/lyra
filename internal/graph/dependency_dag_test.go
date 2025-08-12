package graph

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/errors"
)

func TestDependencyDAGGetExecutionLevels(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name          string
		dependencies  map[string][]string
		expected      [][]string
		expectError   bool
		errorContains error
	}{
		{
			name: "single node no dependencies",
			dependencies: map[string][]string{
				"nodeA": {},
			},
			expected:    [][]string{{"nodeA"}},
			expectError: false,
		},
		{
			name: "two independent nodes",
			dependencies: map[string][]string{
				"nodeA": {},
				"nodeB": {},
			},
			expected:    [][]string{{"nodeA", "nodeB"}},
			expectError: false,
		},
		{
			name: "simple chain two nodes",
			dependencies: map[string][]string{
				"nodeA": {},
				"nodeB": {"nodeA"},
			},
			expected:    [][]string{{"nodeA"}, {"nodeB"}},
			expectError: false,
		},
		{
			name: "simple chain three nodes",
			dependencies: map[string][]string{
				"nodeA": {},
				"nodeB": {"nodeA"},
				"nodeC": {"nodeB"},
			},
			expected:    [][]string{{"nodeA"}, {"nodeB"}, {"nodeC"}},
			expectError: false,
		},
		{
			name: "diamond pattern",
			dependencies: map[string][]string{
				"nodeA": {},
				"nodeB": {},
				"nodeC": {"nodeA", "nodeB"},
				"nodeD": {"nodeC"},
			},
			expected:    [][]string{{"nodeA", "nodeB"}, {"nodeC"}, {"nodeD"}},
			expectError: false,
		},
		{
			name: "complex scenario",
			dependencies: map[string][]string{
				"fetchUser":    {},
				"fetchOrders":  {},
				"createReport": {"fetchUser", "fetchOrders"},
				"processData":  {"createReport"},
				"sendEmail":    {"createReport"},
				"notify":       {"processData", "sendEmail"},
			},
			expected: [][]string{
				{"fetchUser", "fetchOrders"},
				{"createReport"},
				{"processData", "sendEmail"},
				{"notify"},
			},
			expectError: false,
		},
		{
			name: "large fan out",
			dependencies: map[string][]string{
				"root":   {},
				"child1": {"root"},
				"child2": {"root"},
				"child3": {"root"},
				"child4": {"root"},
				"child5": {"root"},
			},
			expected:    [][]string{{"root"}, {"child1", "child2", "child3", "child4", "child5"}},
			expectError: false,
		},
		{
			name: "large fan in",
			dependencies: map[string][]string{
				"source1":   {},
				"source2":   {},
				"source3":   {},
				"source4":   {},
				"source5":   {},
				"collector": {"source1", "source2", "source3", "source4", "source5"},
			},
			expected: [][]string{
				{"source1", "source2", "source3", "source4", "source5"},
				{"collector"},
			},
			expectError: false,
		},
		{
			name: "mixed topology chains and diamonds",
			dependencies: map[string][]string{
				// Chain 1: A -> B -> C
				"nodeA": {},
				"nodeB": {"nodeA"},
				"nodeC": {"nodeB"},

				// Diamond: D,E -> F -> G
				"nodeD": {},
				"nodeE": {},
				"nodeF": {"nodeD", "nodeE"},
				"nodeG": {"nodeF"},

				// Cross dependency: H depends on both chain and diamond
				"nodeH": {"nodeC", "nodeG"},
			},
			expected: [][]string{
				{"nodeA", "nodeD", "nodeE"},
				{"nodeB", "nodeF"},
				{"nodeC", "nodeG"},
				{"nodeH"},
			},
			expectError: false,
		},
		{
			name: "simple cycle two nodes",
			dependencies: map[string][]string{
				"nodeA": {"nodeB"},
				"nodeB": {"nodeA"},
			},
			expected:      nil,
			expectError:   true,
			errorContains: errors.ErrCyclicDependency,
		},
		{
			name: "simple cycle three nodes",
			dependencies: map[string][]string{
				"nodeA": {"nodeB"},
				"nodeB": {"nodeC"},
				"nodeC": {"nodeA"},
			},
			expected:      nil,
			expectError:   true,
			errorContains: errors.ErrCyclicDependency,
		},
		{
			name: "self dependency",
			dependencies: map[string][]string{
				"nodeA": {"nodeA"},
			},
			expected:      nil,
			expectError:   true,
			errorContains: errors.ErrCyclicDependency,
		},
		{
			name: "cycle in complex graph",
			dependencies: map[string][]string{
				"nodeA": {},
				"nodeB": {"nodeA", "nodeE"},
				"nodeC": {"nodeB"},
				"nodeD": {"nodeC"},
				"nodeE": {"nodeD", "nodeB"}, // Creates cycle: B->C->D->E->B but E also needs B
			},
			expected:      nil,
			expectError:   true,
			errorContains: errors.ErrCyclicDependency,
		},
		{
			name: "missing dependency reference",
			dependencies: map[string][]string{
				"nodeA": {"nonExistentNode"},
			},
			expected:      nil,
			expectError:   true,
			errorContains: errors.ErrMissingDependency,
		},
		{
			name:         "empty graph",
			dependencies: map[string][]string{},
			expected:     [][]string{},
			expectError:  false,
		},
		{
			name: "multiple disconnected chains",
			dependencies: map[string][]string{
				"chainA1": {},
				"chainA2": {"chainA1"},
				"chainB1": {},
				"chainB2": {"chainB1"},
				"chainB3": {"chainB2"},
			},
			expected: [][]string{
				{"chainA1", "chainB1"},
				{"chainA2", "chainB2"},
				{"chainB3"},
			},
			expectError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dag := NewDependencyDAG(tc.dependencies)

			levels, err := dag.GetExecutionLevels()

			if tc.expectError {
				require.ErrorIs(t, err, tc.errorContains)
				require.Nil(t, levels)
			} else {
				require.NoError(t, err)
				for i := range len(tc.expected) {
					require.ElementsMatch(t, tc.expected[i], levels[i])
				}
			}
		})
	}
}
