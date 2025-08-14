package lyra

import (
	"context"
	stderr "errors"
	"reflect"
	"sync"

	"github.com/sourabh-kumar2/lyra/errors"
	"github.com/sourabh-kumar2/lyra/internal"
	"github.com/sourabh-kumar2/lyra/internal/graph"
)

// Lyra coordinates dependent tasks that can run concurrently when possible,
// with compile-time type safety for result passing between tasks.
// It replaces manual sync.WaitGroup and channel coordination with a clean, fluent API.
//
// The zero value is not usable; create instances with New().
type Lyra struct {
	mu    sync.RWMutex
	tasks map[string]*internal.Task
	error error
}

// New creates a new Lyra instance for building and executing DAGs.
//
// Example:
//
//	l := lyra.New()
//	l.Do("task1", taskFunc1, lyra.UseRun("input"))
//	l.Do("task2", taskFunc2, lyra.Use("task1"))
//	results, err := l.Run(ctx, map[string]any{"input": "value"})
func New() *Lyra {
	return &Lyra{
		tasks: make(map[string]*internal.Task),
	}
}

// Do adds a task to the DAG with the specified function and input specifications.
//
// The taskID must be unique within the DAG and will be used to reference this
// task's results in other tasks or in the final results.
//
// The fn parameter must be a function with one of these signatures:
//   - func(context.Context) error
//   - func(context.Context) (ResultType, error)
//   - func(context.Context, input1, input2, ...) (ResultType, error)
//
// Input specifications define where each parameter (after context) gets its value:
//   - Use("taskID") - use entire result from another task
//   - Use("taskID", "field") - use specific field from task result
//   - UseRun("key") - use value from runtime inputs map
//
// Returns the same Lyra instance for method chaining.
//
// Example:
//
//	l.Do("fetchUser", fetchUserFunc, lyra.UseRun("userID"))
//	l.Do("processUser", processFunc, lyra.Use("fetchUser", "Name"))
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
//
// The method validates the DAG structure, detects cycles, and executes tasks
// in the optimal order with maximum concurrency. Tasks with no dependencies
// between them run in parallel.
//
// The runInputs map provides initial values that can be referenced by tasks
// using UseRun() input specifications.
//
// Returns a Result object containing all task outputs, or an error if:
//   - The DAG contains cycles
//   - Dependencies reference non-existent tasks
//   - Parameter types don't match between tasks
//   - Any task function returns an error
//
// Example:
//
//	results, err := l.Run(ctx, map[string]any{
//		"userID": 123,
//		"apiKey": "secret",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	user, _ := results.Get("fetchUser")
func (l *Lyra) Run(ctx context.Context, runInputs map[string]any) (*Result, error) {
	if l.error != nil {
		return nil, errors.Wrapf(l.error, "build error")
	}

	result := l.initialiseResult(runInputs)
	stages, err := l.getStages()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get stages")
	}

	err = l.process(ctx, stages, result)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to process stages")
	}

	return result, nil
}

func (*Lyra) initialiseResult(runInputs map[string]any) *Result {
	result := NewResult()
	for taskID, input := range runInputs {
		result.set(taskID, input)
	}
	return result
}

func (l *Lyra) getStages() ([][]string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	taskGraph := map[string][]string{}
	for taskID, task := range l.tasks {
		taskGraph[taskID] = task.GetDependencies()
	}

	stages, err := graph.NewDependencyDAG(taskGraph).GetExecutionLevels()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build graph")
	}
	return stages, nil
}

func (l *Lyra) process(ctx context.Context, stages [][]string, result *Result) error {
	for _, stage := range stages {
		err := l.executeStage(ctx, stage, result)
		if err != nil {
			return errors.Wrapf(err, "execute stage")
		}
	}
	return nil
}

func (l *Lyra) executeStage(ctx context.Context, stage []string, result *Result) error {
	if len(stage) == 1 {
		return l.executeTask(ctx, stage[0], result) // Single task - no need for goroutines
	}
	// Multiple tasks - execute concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(stage))

	for _, taskID := range stage {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := l.executeTask(ctx, id, result); err != nil {
				errChan <- errors.Wrapf(err, "task %q failed", id)
			}
		}(taskID)
	}

	wg.Wait()
	close(errChan)

	//nolint:prealloc // pre-allocating is not required.
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		//nolint:wrapcheck // stderr points to standard errors.
		return stderr.Join(errs...)
	}

	return nil
}

func (l *Lyra) executeTask(ctx context.Context, taskID string, result *Result) error {
	l.mu.RLock()
	task := l.tasks[taskID]
	l.mu.RUnlock()

	args, err := resolveInputs(ctx, task, result)
	if err != nil {
		return errors.Wrapf(err, "input resolution failed")
	}

	values := reflect.ValueOf(task.GetFunction()).Call(args)

	if len(values) == 2 { // (result, error)
		if !values[1].IsNil() {
			// revive:disable-next-line:unchecked-type-assertion // It's always error
			err, _ = values[1].Interface().(error)
			return err
		}
		result.set(taskID, values[0].Interface())
	} else if !values[0].IsNil() { // just (error)
		// revive:disable-next-line:unchecked-type-assertion // It's always error
		err, _ = values[0].Interface().(error)
		return err
	}

	return nil
}
