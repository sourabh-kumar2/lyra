package lyra

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/errors"
)

func TestNewResult(t *testing.T) {
	t.Parallel()

	r := NewResult()

	require.NotNil(t, r)
	require.NotNil(t, r.data, "NewResults() did not initialize data map")
	require.Empty(t, r.data, "NewResults() data map not empty")
}

func TestResultsSet(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name     string
		taskID   string
		value    any
		expected any
	}{
		{
			name:     "set string",
			taskID:   "task1",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "set int",
			taskID:   "task2",
			value:    42,
			expected: 42,
		},
		{
			name:     "set struct",
			taskID:   "task3",
			value:    testStruct{Name: "test", ID: 1},
			expected: testStruct{Name: "test", ID: 1},
		},
		{
			name:     "set slice",
			taskID:   "task4",
			value:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "set nil",
			taskID:   "task5",
			value:    nil,
			expected: nil,
		},
		{
			name:     "overwrite existing",
			taskID:   "task1",
			value:    "world",
			expected: "world",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := NewResult()

			// First set for "overwrite existing" test
			if tc.name == "overwrite existing" {
				r.set("task1", "original")
			}

			r.set(tc.taskID, tc.value)
			got, err := r.Get(tc.taskID)

			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestResultsGet(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name    string
		setup   func(*Result)
		taskID  string
		want    any
		wantErr error
	}{
		{
			name:    "empty results",
			setup:   func(r *Result) {},
			taskID:  "task1",
			want:    nil,
			wantErr: errors.ErrTaskNotFound,
		},
		{
			name:    "non-existent task",
			setup:   func(r *Result) { r.set("other", "value") },
			taskID:  "task1",
			want:    nil,
			wantErr: errors.ErrTaskNotFound,
		},
		{
			name: "existing string task",
			setup: func(r *Result) {
				r.set("task1", "hello")
			},
			taskID:  "task1",
			want:    "hello",
			wantErr: nil,
		},
		{
			name: "existing int task",
			setup: func(r *Result) {
				r.set("task1", 42)
			},
			taskID:  "task1",
			want:    42,
			wantErr: nil,
		},
		{
			name: "existing struct task",
			setup: func(r *Result) {
				r.set("task1", testStruct{Name: "test", ID: 1})
			},
			taskID:  "task1",
			want:    testStruct{Name: "test", ID: 1},
			wantErr: nil,
		},
		{
			name: "existing slice task",
			setup: func(r *Result) {
				r.set("task1", []string{"a", "b", "c"})
			},
			taskID:  "task1",
			want:    []string{"a", "b", "c"},
			wantErr: nil,
		},
		{
			name: "existing nil value",
			setup: func(r *Result) {
				r.set("task1", nil)
			},
			taskID:  "task1",
			want:    nil,
			wantErr: nil,
		},
		{
			name:    "empty task ID",
			setup:   func(r *Result) {},
			taskID:  "",
			want:    nil,
			wantErr: errors.ErrTaskNotFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := NewResult()
			tc.setup(r)

			got, err := r.Get(tc.taskID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func TestResultsGetFromZeroValue(t *testing.T) {
	t.Parallel()

	var r Result // Zero value - data will be nil

	got, err := r.Get("task1")

	require.Nil(t, got)
	require.ErrorIs(t, err, errors.ErrTaskNotFound)
}

func TestResultsSetLazyInit(t *testing.T) {
	t.Parallel()

	var r Result // Zero value - data will be nil

	r.set("task1", "hello")

	require.NotNil(t, r.data)

	got, err := r.Get("task1")
	require.NoError(t, err)
	require.Equal(t, "hello", got)
}

func TestResultsConcurrency(t *testing.T) {
	t.Parallel()

	r := NewResult()

	// Pre-populate some data
	r.set("existing", "value")

	var wg sync.WaitGroup
	numGoroutines := 10

	// Start readers
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			val, err := r.Get("existing")
			require.NoErrorf(t, err, "reader goroutine %d: unexpected error reading existing", id)
			require.Equalf(t, "value", val, "reader goroutine %d", id)

			_, err = r.Get("non-existing")
			require.ErrorIsf(t, err, errors.ErrTaskNotFound, "reader goroutine %d", id)
		}(i)
	}

	// Start writers (writing to different keys)
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			taskID := fmt.Sprintf("task_%d", id)
			value := fmt.Sprintf("value_%d", id)

			r.set(taskID, value)

			// Verify write
			got, err := r.Get(taskID)
			require.NoErrorf(t, err, "writer goroutine %d: unexpected error reading task", id)
			require.Equalf(t, value, got, "writer goroutine %d", id)
		}(i)
	}

	wg.Wait()

	require.Len(t, r.data, numGoroutines+1)
}

type testStruct struct {
	Name string
	ID   int
}
