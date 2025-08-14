package lyra

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/internal"
)

func TestUse(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name      string
		source    string
		fieldPath []string
		expected  internal.InputSpec
	}{
		{
			name:   "simple task result",
			source: "fetchUser",
			expected: internal.InputSpec{
				Type:   internal.TaskResultInputSpec,
				Source: "fetchUser",
			},
		},
		{
			name:      "task result with field",
			source:    "fetchUser",
			fieldPath: []string{"ID"},
			expected: internal.InputSpec{
				Type:   internal.TaskResultInputSpec,
				Source: "fetchUser",
				Field:  []string{"ID"},
			},
		},
		{
			name:      "nested field path",
			source:    "fetchUser",
			fieldPath: []string{"Address", "Street"},
			expected: internal.InputSpec{
				Type:   internal.TaskResultInputSpec,
				Source: "fetchUser",
				Field:  []string{"Address", "Street"},
			},
		},
		{
			name:      "deep nested field path",
			source:    "fetchUser",
			fieldPath: []string{"Profile", "Settings", "Theme", "Color"},
			expected: internal.InputSpec{
				Type:   internal.TaskResultInputSpec,
				Source: "fetchUser",
				Field:  []string{"Profile", "Settings", "Theme", "Color"},
			},
		},
		{
			name:   "empty source",
			source: "",
			expected: internal.InputSpec{
				Type:   internal.TaskResultInputSpec,
				Source: "",
			},
		},
		{
			name:      "empty field path elements",
			source:    "fetchUser",
			fieldPath: []string{"", "ID", ""},
			expected: internal.InputSpec{
				Type:   internal.TaskResultInputSpec,
				Source: "fetchUser",
				Field:  []string{"", "ID", ""},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result := Use(tc.source, tc.fieldPath...)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUseRun(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name      string
		source    string
		fieldPath []string
		expected  internal.InputSpec
	}{
		{
			name:   "simple literal",
			source: "user_id",
			expected: internal.InputSpec{
				Type:   internal.RuntimeInputSpec,
				Source: "user_id",
			},
		},
		{
			name:      "struct with field",
			source:    "user",
			fieldPath: []string{"ID"},
			expected: internal.InputSpec{
				Type:   internal.RuntimeInputSpec,
				Source: "user",
				Field:  []string{"ID"},
			},
		},
		{
			name:      "nested field path",
			source:    "user",
			fieldPath: []string{"Address", "Street"},
			expected: internal.InputSpec{
				Type:   internal.RuntimeInputSpec,
				Source: "user",
				Field:  []string{"Address", "Street"},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result := UseRun(tc.source, tc.fieldPath...)
			require.Equal(t, tc.expected, result)
		})
	}
}
