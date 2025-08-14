package internal

type inputSpecType = int

const (
	// RuntimeInputSpec defines for user input.
	RuntimeInputSpec inputSpecType = iota

	// TaskResultInputSpec defines for task output used as input.
	TaskResultInputSpec inputSpecType = iota
)

// InputSpec specifies how to get input for a task parameter.
// This type is used internally and created by lyra.Use() and lyra.UseRun() functions.
//
// Do not create InputSpec instances directly; use the provided helper functions.
type InputSpec struct {
	Type   inputSpecType // Type Distinguishes between runtime and task dependency inputs
	Source string        // Source task ID or runtime key
	Field  []string      // Field Optional nested field path
}
