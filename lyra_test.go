package lyra

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/errors"
	"github.com/sourabh-kumar2/lyra/internal"
)

func TestNew(t *testing.T) {
	t.Parallel()

	l := New()
	require.NotNil(t, l)
	require.NotNil(t, l.tasks)
	require.Len(t, l.tasks, 0)
	require.Nil(t, l.error)
}

func TestDo(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name              string
		tasks             []testTask
		expectedTaskCount int
		expectedErr       error
	}{
		{
			name: "single task",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
			},
			expectedTaskCount: 1,
		},
		{
			name: "multiple tasks",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
				{
					id:         "task-2",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
				{
					id:         "task-3",
					fn:         anotherValidTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
				{
					id:         "task-4",
					fn:         dependentTask,
					inputSpecs: []internal.InputSpec{Use("task-2")},
				},
			},
			expectedTaskCount: 4,
		},
		{
			name: "valid task function with empty inputSpecs",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: nil,
				},
			},
			expectedTaskCount: 1,
		},
		{
			name: "invalid task function with empty inputSpecs",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         invalidTask,
					inputSpecs: nil,
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrMustHaveAtLeastContext,
		},
		{
			name: "empty id",
			tasks: []testTask{
				{
					id:         "",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskIDCannotBeEmpty,
		},
		{
			name: "empty id with whitespace",
			tasks: []testTask{
				{
					id:         "     ",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskIDCannotBeEmpty,
		},
		{
			name: "duplicate task id",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         anotherValidTask,
					inputSpecs: []internal.InputSpec{UseRun("orderID")},
				},
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: nil,
				},
			},
			expectedTaskCount: 1,
			expectedErr:       errors.ErrDuplicateTask,
		},
		{
			name: "invalid task function",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         invalidTask,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrMustHaveAtLeastContext,
		},
		{
			name: "wrong input count less than expected",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskParamCountMismatch,
		},
		{
			name: "wrong input count greater than expected",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskParamCountMismatch,
		},
		{
			name: "nil function",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         nil,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrMustBeFunction,
		},
		{
			name: "adding tasks even after error",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
				{
					id:         "task-2",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
				{
					id:         "task-3",
					fn:         anotherValidTask,
					inputSpecs: []internal.InputSpec{UseRun("orderID"), UseRun("userID")},
				},
				{
					id:         "task-4",
					fn:         dependentTask,
					inputSpecs: []internal.InputSpec{Use("task-2")},
				},
			},
			expectedTaskCount: 3,
			expectedErr:       errors.ErrTaskParamCountMismatch,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := New()
			for _, task := range tc.tasks {
				l.Do(task.id, task.fn, task.inputSpecs...)
			}
			require.ErrorIs(t, l.error, tc.expectedErr)
			require.Len(t, l.tasks, tc.expectedTaskCount)
		})
	}
}

func TestDoConcurrency(t *testing.T) {
	t.Parallel()

	l := New()

	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			l.Do(fmt.Sprintf("task-%d", id), validTaskWithNoInput)
		}(i)
	}

	wg.Wait()
	require.Len(t, l.tasks, 10)
}

func TestRunEmptyDAG(t *testing.T) {
	t.Parallel()

	l := New()
	result, err := l.Run(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestRunEmptyDAGWithRunInputs(t *testing.T) {
	t.Parallel()

	runInputs := map[string]any{
		"userID":  123,
		"orderID": 456,
	}

	l := New()
	result, err := l.Run(context.Background(), runInputs)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, result.data, runInputs)
}

type user struct {
	ID string
}

func validTask(ctx context.Context, userID string) (user, error) {
	return user{}, nil
}

func validTaskWithNoInput(ctx context.Context) error {
	return nil
}

func anotherValidTask(ctx context.Context, orderID string) error {
	return nil
}

func dependentTask(ctx context.Context, user user) error {
	return nil
}

func invalidTask() {
}

type testTask struct {
	id         string
	fn         any
	inputSpecs []internal.InputSpec
}
