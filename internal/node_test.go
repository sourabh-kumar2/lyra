package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func validTaskFunc(ctx context.Context, userID string) (string, error) {
	return "result", nil
}

func TestNewTask(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name       string
		id         string
		fn         any
		inputSpecs []InputSpec
		wantErr    bool
		errType    error
	}{
		{
			name:    "valid node creation",
			id:      "testTask",
			fn:      validTaskFunc,
			wantErr: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			task, err := NewTask(tc.id, tc.fn, tc.inputSpecs)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, task)
				if tc.errType != nil {
					require.ErrorIs(t, err, tc.errType)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				require.Equal(t, tc.id, task.id)
			}
		})
	}
}
