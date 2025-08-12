package lyra

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUse(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name      string
		source    string
		fieldPath []string
		expected  InputSpec
	}{
		{
			name:   "simple task result",
			source: "fetchUser",
			expected: InputSpec{
				Type:   taskResultInputSpec,
				Source: "fetchUser",
				Field:  "",
			},
		},
		{
			name:      "task result with field",
			source:    "fetchUser",
			fieldPath: []string{"ID"},
			expected: InputSpec{
				Type:   taskResultInputSpec,
				Source: "fetchUser",
				Field:  "ID",
			},
		},
		{
			name:      "nested field path",
			source:    "fetchUser",
			fieldPath: []string{"Address", "Street"},
			expected: InputSpec{
				Type:   taskResultInputSpec,
				Source: "fetchUser",
				Field:  "Address.Street",
			},
		},
		{
			name:      "deep nested field path",
			source:    "fetchUser",
			fieldPath: []string{"Profile", "Settings", "Theme", "Color"},
			expected: InputSpec{
				Type:   taskResultInputSpec,
				Source: "fetchUser",
				Field:  "Profile.Settings.Theme.Color",
			},
		},
		{
			name:   "empty source",
			source: "",
			expected: InputSpec{
				Type:   taskResultInputSpec,
				Source: "",
				Field:  "",
			},
		},
		{
			name:      "empty field path elements",
			source:    "fetchUser",
			fieldPath: []string{"", "ID", ""},
			expected: InputSpec{
				Type:   taskResultInputSpec,
				Source: "fetchUser",
				Field:  "ID",
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
		expected  InputSpec
	}{
		{
			name:   "simple literal",
			source: "user_id",
			expected: InputSpec{
				Type:   runtimeInputSpec,
				Source: "user_id",
				Field:  "",
			},
		},
		{
			name:      "struct with field",
			source:    "user",
			fieldPath: []string{"ID"},
			expected: InputSpec{
				Type:   runtimeInputSpec,
				Source: "user",
				Field:  "ID",
			},
		},
		{
			name:      "nested field path",
			source:    "user",
			fieldPath: []string{"Address", "Street"},
			expected: InputSpec{
				Type:   runtimeInputSpec,
				Source: "user",
				Field:  "Address.Street",
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
