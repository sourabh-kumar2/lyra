package lyra

// InputSpec specifies how to get input for a task parameter.
type InputSpec struct{}

// Use creates an InputSpec for task results or runtime inputs
//
// Examples:
//
//	Use("fetchUser")           -> Task result from "fetchUser"
//	Use("fetchUser", "ID")     -> Field "ID" from "fetchUser" result
//	Use("fetchUser", "Address", "Street") -> Nested field "Address.Street"
//	Use("runtime", "userID")   -> Runtime input "userID"
func Use(source string, fieldPath ...string) InputSpec {
	_ = source
	_ = fieldPath
	return InputSpec{}
}
