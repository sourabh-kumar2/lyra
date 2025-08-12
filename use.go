package lyra

import (
	"strings"

	"github.com/sourabh-kumar2/lyra/internal"
)

// Use creates an InputSpec for task results inputs
//
// Examples:
//
//	Use("fetchUser")           -> Task result from "fetchUser"
//	Use("fetchUser", "ID")     -> Field "ID" from "fetchUser" result
//	Use("fetchUser", "Address", "Street") -> Nested field "Address.Street"
func Use(source string, fieldPath ...string) internal.InputSpec {
	field := ""
	if len(fieldPath) > 0 {
		field = strings.Trim(strings.Join(fieldPath, "."), ".")
	}
	return internal.InputSpec{
		Type:   internal.TaskResultInputSpec,
		Source: source,
		Field:  field,
	}
}

// UseRun creates an InputSpec for taking inputs from Run.
//
// Examples:
//
// UseRun("user_id")           -> user_id from Run(ctx, map[string]any{"user_id": 123}).
func UseRun(source string, fieldPath ...string) internal.InputSpec {
	it := Use(source, fieldPath...)
	it.Type = internal.RuntimeInputSpec
	return it
}
