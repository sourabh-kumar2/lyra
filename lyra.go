package lyra

import (
	"context"
	"reflect"
	"sync"

	"github.com/sourabh-kumar2/lyra/errors"
	"github.com/sourabh-kumar2/lyra/internal"
	"github.com/sourabh-kumar2/lyra/internal/graph"
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
func (l *Lyra) Run(ctx context.Context, runInputs map[string]any) (*Result, error) {
	if l.error != nil {
		return nil, errors.Wrapf(l.error, "build error")
	}

	result := l.initialiseResult(runInputs)
	stages, err := l.getStages()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get stages")
	}

	l.process(ctx, stages, result)

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

func (l *Lyra) process(ctx context.Context, stages [][]string, result *Result) {
	for _, stage := range stages {
		_ = l.executeStage(ctx, stage, result)
	}
}

func (l *Lyra) executeStage(ctx context.Context, stage []string, result *Result) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, taskID := range stage {
		task := l.tasks[taskID]
		args, err := resolveInputs(ctx, task, result)
		if err != nil {
			return errors.Wrapf(err, "failed to resolve task %q", taskID)
		}
		values := reflect.ValueOf(task.GetFunction()).Call(args)
		if len(values) == 2 {
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
	}
	return nil
}
