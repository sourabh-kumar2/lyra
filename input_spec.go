package lyra

import "strings"

type inputSpecType = int

const (
	runtimeInputSpec    inputSpecType = iota
	taskResultInputSpec inputSpecType = iota
)

// InputSpec specifies how to get input for a task parameter.
type InputSpec struct {
	Type   inputSpecType // Type is required to distinguish between runtime params and task dependency params.
	Source string
	Field  string
}

// Use creates an InputSpec for task results inputs
//
// Examples:
//
//	Use("fetchUser")           -> Task result from "fetchUser"
//	Use("fetchUser", "ID")     -> Field "ID" from "fetchUser" result
//	Use("fetchUser", "Address", "Street") -> Nested field "Address.Street"
func Use(source string, fieldPath ...string) InputSpec {
	field := ""
	if len(fieldPath) > 0 {
		field = strings.Trim(strings.Join(fieldPath, "."), ".")
	}
	return InputSpec{
		Type:   taskResultInputSpec,
		Source: source,
		Field:  field,
	}
}

// UseRun creates an InputSpec for taking inputs from Run.
//
// Examples:
//
// UseRun("user_id")           -> user_id from Run(ctx, map[string]any{"user_id": 123}).
func UseRun(source string, fieldPath ...string) InputSpec {
	it := Use(source, fieldPath...)
	it.Type = runtimeInputSpec
	return it
}
