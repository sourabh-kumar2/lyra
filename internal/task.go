package internal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourabh-kumar2/lyra/errors"
)

// Task represents a single task node in the DAG with its function and dependencies.
// This type is used internally and should not be created directly.
type Task struct {
	id         string
	fn         any
	fnInfo     *functionInfo
	inputSpecs []InputSpec
}

// NewTask creates a task node with validation.
// This function is used internally and validates that:
//   - The task ID is not empty
//   - The function signature is valid
//   - The number of input specs matches function parameters
//
// Returns an error if validation fails.
func NewTask(id string, fn any, inputSpecs []InputSpec) (*Task, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.ErrTaskIDCannotBeEmpty
	}
	fnInfo, err := analyzeFunctionSignature(fn)
	if err != nil {
		return nil, fmt.Errorf("invalid function for task %q: %w", id, err)
	}
	if len(inputSpecs) != len(fnInfo.inputTypes)-1 {
		return nil, errors.Wrapf(
			errors.ErrTaskParamCountMismatch,
			"invalid number of input specs for task %q, want: %d, got: %d",
			id,
			len(fnInfo.inputTypes),
			len(inputSpecs)+1,
		)
	}
	return &Task{
		id:         id,
		fn:         fn,
		inputSpecs: inputSpecs,
		fnInfo:     fnInfo,
	}, nil
}

// GetDependencies returns the task IDs that this task depends on.
// Only returns dependencies from TaskResultInputSpec types (lyra.Use() calls),
// not runtime inputs (lyra.UseRun() calls).
func (t *Task) GetDependencies() []string {
	deps := make([]string, 0)
	for _, spec := range t.inputSpecs {
		if spec.Type == TaskResultInputSpec {
			deps = append(deps, spec.Source)
		}
	}
	return deps
}

// GetInputParams returns the input specifications and parameter types for this task.
// Used internally during execution to resolve parameter values.
func (t *Task) GetInputParams() (specs []InputSpec, types []reflect.Type) {
	return t.inputSpecs, t.fnInfo.inputTypes
}

// GetFunction returns the callable function for this task.
// Used internally during task execution.
func (t *Task) GetFunction() any {
	return t.fn
}

// GetOutputParams returns the output type if the function returns a result.
// Returns nil if the function only returns an error.
func (t *Task) GetOutputParams() reflect.Type {
	return t.fnInfo.outputType
}

// GetID returns the task's unique identifier.
func (t *Task) GetID() string {
	return t.id
}
