package lyra

import "context"

// Lyra coordinates dependent tasks that can run concurrently when possible,
// with compile-time type safety for result passing between tasks.
// It replaces manual sync.WaitGroup and channel
// coordination with a clean, fluent API.
type Lyra struct{}

// New creates a new Lyra instance for building and executing DAGs.
//
//	l := lyra.New()
//	l.Do("task1", taskFunc1).Do("task2", taskFunc2).After("task1")
func New() *Lyra {
	return &Lyra{}
}

// Do adds a task to the DAG and returns a TaskBuilder for chaining.
func (l *Lyra) Do(taskID string, fn any, inputs ...InputSpec) *Lyra {
	_ = taskID
	_ = fn
	_ = inputs
	return l
}

// Run executes the DAG with the provided runtime inputs.
func (*Lyra) Run(_ context.Context, _ map[string]any) (*Result, error) {
	return &Result{}, nil
}
