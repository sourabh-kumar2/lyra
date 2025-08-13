package lyra

import (
	"context"
	"sync"

	"github.com/sourabh-kumar2/lyra/errors"
	"github.com/sourabh-kumar2/lyra/internal"
)

// Lyra coordinates dependent tasks that can run concurrently when possible,
// with compile-time type safety for result passing between tasks.
// It replaces manual sync.WaitGroup and channel
// coordination with a clean, fluent API.
type Lyra struct {
	mu    sync.RWMutex
	tasks map[string]*internal.Task
	error error
}

// New creates a new Lyra instance for building and executing DAGs.
//
//	l := lyra.New()
//	l.Do("task1", taskFunc1).Do("task2", taskFunc2).After("task1")
func New() *Lyra {
	return &Lyra{
		tasks: make(map[string]*internal.Task),
	}
}

// Do adds a task to the DAG and returns a TaskBuilder for chaining.
func (l *Lyra) Do(taskID string, fn any, inputs ...internal.InputSpec) *Lyra {
	l.mu.Lock()
	defer l.mu.Unlock()

	task, err := internal.NewTask(taskID, fn, inputs)
	if err != nil {
		l.error = errors.Wrapf(err, "failed to add task %q", taskID)
		return l
	}
	if _, exists := l.tasks[taskID]; exists {
		l.error = errors.Wrapf(errors.ErrDuplicateTask, "failed to add task %q", taskID)
		return l
	}
	l.tasks[taskID] = task
	return l
}

// Run executes the DAG with the provided runtime inputs.
func (*Lyra) Run(_ context.Context, _ map[string]any) (*Result, error) {
	return &Result{}, nil
}
