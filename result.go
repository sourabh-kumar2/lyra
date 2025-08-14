package lyra

import (
	"sync"

	"github.com/sourabh-kumar2/lyra/errors"
)

// Result holds the results of DAG execution in a thread-safe manner.
// Results are stored as interface{} and require type assertion when retrieved.
//
// The zero value is not usable; Result instances are created by Lyra.Run().
type Result struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewResult creates a new Result instance for storing task execution results.
// This is primarily used internally by Lyra, but can be useful for testing.
func NewResult() *Result {
	return &Result{
		data: make(map[string]any),
	}
}

// Get retrieves the result for the specified task ID.
//
// Returns the task result and nil error if found, or nil and ErrTaskNotFound
// if the task ID doesn't exist in the results.
//
// The returned value requires type assertion to the expected type:
//
//	user, err := results.Get("fetchUser")
//	if err != nil {
//		return err
//	}
//	typedUser := user.(User) // Type assertion required
//
// For safer type handling, consider storing results in typed variables
// immediately after retrieval.
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
