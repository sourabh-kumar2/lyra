package errors

import (
	"errors"
	"fmt"
)

// ErrMustBeFunction is returned when the provided value is not a function.
var ErrMustBeFunction = errors.New("must be a function")

// ErrMustHaveAtLeastContext is returned when function has no parameters.
var ErrMustHaveAtLeastContext = errors.New("must have at least one parameter (context.Context)")

// ErrFirstParamMustBeContext is returned when first parameter is not context.Context.
var ErrFirstParamMustBeContext = errors.New("first parameter must be context.Context")

// ErrMustReturnAtLeastError is returned when function has no return values.
var ErrMustReturnAtLeastError = errors.New("must return at least error")

// ErrSingleReturnMustBeError is returned when single return value doesn't implement error.
var ErrSingleReturnMustBeError = errors.New("single return value must implement error interface")

// ErrSecondReturnMustBeError is returned when second return value doesn't implement error.
var ErrSecondReturnMustBeError = errors.New("second return value must implement error interface")

// ErrTooManyReturnValues is returned when function returns more than 2 values.
var ErrTooManyReturnValues = errors.New("must return 1 or 2 values")

// ErrVariadicNotSupported is returned when function uses variadic parameters.
var ErrVariadicNotSupported = errors.New("variadic functions are not supported")

// ErrTaskIDCannotBeEmpty is returned when the task id is empty.
var ErrTaskIDCannotBeEmpty = errors.New("task id must not be empty")

// ErrTaskParamCountMismatch is returned when the task has params mismatch.
var ErrTaskParamCountMismatch = errors.New("task params count mismatch")

// ErrCyclicDependency is returned when DAG contains circular dependencies.
var ErrCyclicDependency = errors.New("cyclic dependency detected")

// ErrMissingDependency is returned when referenced dependency doesn't exist.
var ErrMissingDependency = errors.New("dependency not found")

// ErrInvalidInput is returned when task has invalid input spec.
var ErrInvalidInput = errors.New("invalid input")

// Wrapf returns the wrapped error.
// nolint:err113 // we are wrapping here so needed.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return fmt.Errorf(format, args...)
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
