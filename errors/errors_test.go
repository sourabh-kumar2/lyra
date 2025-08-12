package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var errTest = errors.New("some error occurred")

func TestWrapf(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name   string
		format string
		args   []any
		err    error
		want   string
	}{
		{
			name:   "nil error",
			err:    nil,
			format: "%v",
			args:   []any{"some error occurred"},
			want:   "some error occurred",
		},
		{
			name:   "wrapped error",
			err:    errTest,
			format: "missing dependency node %q not found",
			args:   []any{"nodeA"},
			want:   "missing dependency node \"nodeA\" not found: some error occurred",
		},
		{
			name:   "no args",
			err:    errTest,
			format: "missing dependency",
			args:   []any{},
			want:   "missing dependency: some error occurred",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := Wrapf(tc.err, tc.format, tc.args...).Error()
			require.Equal(t, tc.want, got)
		})
	}
}
