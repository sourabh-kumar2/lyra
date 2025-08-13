package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/errors"
)

func validTaskFunc(ctx context.Context, userID string) (string, error) {
	return "result", nil
}

func multiParamTaskFunc(ctx context.Context, userID string, count int) (string, error) {
	return "result", nil
}

func invalidTaskFunc(userID string) (string, error) { // Missing context
	return "result", nil
}

func noReturnTaskFunc(ctx context.Context) {} // Invalid - no returns

func TestNewTask(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name       string
		id         string
		fn         any
		wantErr    bool
		inputSpecs []InputSpec
		errType    error
	}{
		{
			name: "valid node creation",
			id:   "testTask",
			fn:   validTaskFunc,
			inputSpecs: []InputSpec{
				{
					Type:   RuntimeInputSpec,
					Source: "userID",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty id should fail",
			id:      "",
			fn:      validTaskFunc,
			wantErr: true,
			errType: errors.ErrTaskIDCannotBeEmpty,
		},
		{
			name:    "whitespace id should fail",
			id:      "   ",
			fn:      validTaskFunc,
			wantErr: true,
			errType: errors.ErrTaskIDCannotBeEmpty,
		},
		{
			name:    "nil function should fail",
			id:      "testTask",
			fn:      nil,
			wantErr: true,
			errType: errors.ErrMustBeFunction,
		},
		{
			name:    "invalid function signature should fail",
			id:      "testTask",
			fn:      invalidTaskFunc,
			wantErr: true,
			errType: errors.ErrFirstParamMustBeContext,
		},
		{
			name:    "no return function should fail",
			id:      "testTask",
			fn:      noReturnTaskFunc,
			wantErr: true,
			errType: errors.ErrMustReturnAtLeastError,
		},
		{
			name: "params count mismatch should fail",
			id:   "testTask",
			fn:   multiParamTaskFunc,
			inputSpecs: []InputSpec{
				{
					Type:   RuntimeInputSpec,
					Source: "userID",
				},
				{
					Type:   RuntimeInputSpec,
					Source: "count",
				},
				{
					Type:   RuntimeInputSpec,
					Source: "order",
				},
			},
			wantErr: true,
			errType: errors.ErrTaskParamCountMismatch,
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
