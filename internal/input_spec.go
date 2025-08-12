package internal

type inputSpecType = int

const (
	// RuntimeInputSpec defines for user input.
	RuntimeInputSpec inputSpecType = iota

	// TaskResultInputSpec defines for task output used as input.
	TaskResultInputSpec inputSpecType = iota
)

// InputSpec specifies how to get input for a task parameter.
type InputSpec struct {
	Type   inputSpecType // Type is required to distinguish between runtime params and task dependency params.
	Source string
	Field  string
}
