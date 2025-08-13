package internal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourabh-kumar2/lyra/errors"
)

// Task represents a single task in a DAG.
type Task struct {
	id         string
	fn         any
	fnInfo     *functionInfo
	inputSpecs []InputSpec
}

// NewTask creates a task node.
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

// GetDependencies returns the node dependencies.
func (t *Task) GetDependencies() []string {
	deps := make([]string, 0)
	for _, spec := range t.inputSpecs {
		if spec.Type == TaskResultInputSpec {
			deps = append(deps, spec.Source)
		}
	}
	return deps
}

// GetInputParams returns input params for the calling function.
func (t *Task) GetInputParams() (specs []InputSpec, types []reflect.Type) {
	return t.inputSpecs, t.fnInfo.inputTypes
}

// GetFunction returns the callable function.
func (t *Task) GetFunction() any {
	return t.fn
}

// GetOutputParams returns the output type if it exists.
func (t *Task) GetOutputParams() reflect.Type {
	return t.fnInfo.outputType
}
