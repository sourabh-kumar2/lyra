package internal

// Task represents a single task in a DAG.
type Task struct {
	id         string
	fn         any
	inputSpecs []InputSpec
}

// NewTask creates a task node.
func NewTask(id string, fn any, inputSpecs []InputSpec) (*Task, error) {
	return &Task{
		id:         id,
		fn:         fn,
		inputSpecs: inputSpecs,
	}, nil
}
