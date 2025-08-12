package lyra

import "strings"

// InputSpec specifies how to get input for a task parameter.
type InputSpec struct {
	Source string
	Field  string
}

// Use creates an InputSpec for task results or runtime inputs
//
// Examples:
//
//	Use("fetchUser")           -> Task result from "fetchUser"
//	Use("fetchUser", "ID")     -> Field "ID" from "fetchUser" result
//	Use("fetchUser", "Address", "Street") -> Nested field "Address.Street"
//	Use("runtime", "userID")   -> Runtime input "userID"
func Use(source string, fieldPath ...string) InputSpec {
	field := ""
	if len(fieldPath) > 0 {
		field = strings.Trim(strings.Join(fieldPath, "."), ".")
	}
	return InputSpec{
		Source: source,
		Field:  field,
	}
}
