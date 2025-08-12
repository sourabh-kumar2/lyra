package lyra

import (
	"sync"

	"github.com/sourabh-kumar2/lyra/errors"
)

// Result holds the results of DAG execution.
type Result struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewResult creates a new Result instance for storing task execution results.
func NewResult() *Result {
	return &Result{
		data: make(map[string]any),
	}
}

// Get retrieves the result for the specified task ID.
// Returns the task result and nil error if found, or nil and wrapped errors.ErrTaskNotFound if not
// found.
func (r *Result) Get(taskID string) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, ok := r.data[taskID]
	if !ok {
		return nil, errors.Wrapf(errors.ErrTaskNotFound, "taskID:%s", taskID)
	}
	return data, nil
}

// set stores a result for the given task ID. Initializes internal storage if needed.
func (r *Result) set(taskID string, result any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.data == nil {
		r.data = make(map[string]any)
	}
	r.data[taskID] = result
}
