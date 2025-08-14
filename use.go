package lyra

import (
	"github.com/sourabh-kumar2/lyra/internal"
)

// Use creates an internal.InputSpec for task result inputs with optional nested field access.
//
// This function specifies that a task parameter should receive its value from
// another task's result. Supports accessing nested fields using dot notation.
//
// Examples:
//
//	Use("fetchUser")                    // Entire result from "fetchUser" task
//	Use("fetchUser", "ID")              // Field "ID" from fetchUser result
//	Use("fetchUser", "Address", "Street") // Nested field "Address.Street"
//
// The field path supports accessing:
//   - Struct fields (exported fields only)
//   - Nested structs with multiple levels
//   - Fields through pointer dereference
//
// Returns an InputSpec that can be passed to Lyra.Do().
func Use(source string, fieldPath ...string) internal.InputSpec {
	return internal.InputSpec{
		Type:   internal.TaskResultInputSpec,
		Source: source,
		Field:  fieldPath,
	}
}

// UseRun creates an InputSpec for taking inputs from the Run method's input map.
//
// This function specifies that a task parameter should receive its value from
// the runtime inputs provided to Lyra.Run(). Supports nested field access
// similar to Use().
//
// Examples:
//
//	UseRun("userID")                    // Value of "userID" from run inputs
//	UseRun("config", "Database", "Host") // Nested field "config.Database.Host"
//
// The runtime input map is passed to Run():
//
//	results, err := l.Run(ctx, map[string]any{
//		"userID": 123,
//		"config": DatabaseConfig{...},
//	})
//
// Returns an internal.InputSpec that can be passed to Lyra.Do().
func UseRun(source string, fieldPath ...string) internal.InputSpec {
	it := Use(source, fieldPath...)
	it.Type = internal.RuntimeInputSpec
	return it
}
