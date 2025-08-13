package internal

import (
	"context"
	"reflect"
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

func TestGetDependencies(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name       string
		fn         any
		inputSpecs []InputSpec
		expected   []string
	}{
		{
			name:       "no dependencies",
			inputSpecs: []InputSpec{},
			fn:         func(ctx context.Context) error { return nil },
			expected:   []string{},
		},
		{
			name:       "nil dependency",
			fn:         func(ctx context.Context) error { return nil },
			inputSpecs: nil,
			expected:   []string{},
		},
		{
			name: "one runtime dependency",
			fn:   validTaskFunc,
			inputSpecs: []InputSpec{
				{
					Type:   RuntimeInputSpec,
					Source: "userID",
				},
			},
			expected: []string{},
		},
		{
			name: "one task result dependencies",
			fn:   func(ctx context.Context, user User) error { return nil },
			inputSpecs: []InputSpec{
				{
					Type:   TaskResultInputSpec,
					Source: "fetchUser",
				},
			},
			expected: []string{"fetchUser"},
		},
		{
			name: "multiple runtime dependencies",
			fn:   func(ctx context.Context, user User, userID string, order Order) error { return nil },
			inputSpecs: []InputSpec{
				{
					Type:   TaskResultInputSpec,
					Source: "fetchUser",
				},
				{
					Type:   RuntimeInputSpec,
					Source: "userID",
				},
				{
					Type:   TaskResultInputSpec,
					Source: "fetchOrder",
				},
			},
			expected: []string{"fetchUser", "fetchOrder"},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			task, err := NewTask("test", tc.fn, tc.inputSpecs)
			require.NoError(t, err)
			require.NotNil(t, task)

			deps := task.GetDependencies()
			require.Equal(t, tc.expected, deps)
		})
	}
}

func TestGetInputParams(t *testing.T) {
	t.Parallel()
	inputSpecs := []InputSpec{
		{
			Type:   RuntimeInputSpec,
			Source: "userID",
		},
	}
	task, err := NewTask(
		"id",
		func(ctx context.Context, userID string) error { return nil },
		inputSpecs,
	)

	require.NoError(t, err)
	specs, types := task.GetInputParams()
	require.Equal(t, inputSpecs, specs)
	require.Equal(t, []reflect.Type{
		contextInterface,
		reflect.TypeOf(""),
	}, types)
}

func TestGetFunction(t *testing.T) {
	t.Parallel()
	inputSpecs := []InputSpec{
		{
			Type:   RuntimeInputSpec,
			Source: "userID",
		},
	}
	task, err := NewTask(
		"id",
		func(ctx context.Context, userID string) error { return nil },
		inputSpecs,
	)

	require.NoError(t, err)
	fn := task.GetFunction()
	require.NotNil(t, fn)
}

func TestGetOutputParams(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name     string
		fn       any
		expected reflect.Type
	}{
		{
			name: "no output",
			fn:   func(ctx context.Context) error { return nil },
		},
		{
			name:     "string output",
			fn:       func(ctx context.Context) (string, error) { return "", nil },
			expected: reflect.TypeOf(""),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			task, err := NewTask("id", tc.fn, nil)
			require.NoError(t, err)
			outType := task.GetOutputParams()
			require.Equal(t, tc.expected, outType)
		})
	}
}

func TestGetID(t *testing.T) {
	t.Parallel()
	inputSpecs := []InputSpec{
		{
			Type:   RuntimeInputSpec,
			Source: "userID",
		},
	}
	taskID := "task-1"
	task, err := NewTask(
		taskID,
		func(ctx context.Context, userID string) error { return nil },
		inputSpecs,
	)

	require.NoError(t, err)
	require.Equal(t, taskID, task.GetID())
}
