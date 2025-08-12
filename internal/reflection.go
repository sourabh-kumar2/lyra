package internal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/sourabh-kumar2/lyra/errors"
)

type numberOfOutputs = int

const (
	noOutput             numberOfOutputs = iota
	onlyErrorOutput      numberOfOutputs = iota
	resultAndErrorOutput numberOfOutputs = iota
)

var (
	errorInterface   = reflect.TypeOf((*error)(nil)).Elem()
	contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
)

func analyzeFunctionSignature(fn any) (*functionInfo, error) {
	if fn == nil {
		return nil, errors.ErrMustBeFunction
	}

	fnType := reflect.TypeOf(fn)

	err := validateFunction(fnType)
	if err != nil {
		return nil, fmt.Errorf("invalid function signature: %w", err)
	}

	inputTypes := make([]reflect.Type, fnType.NumIn())
	for i := range fnType.NumIn() {
		inputTypes[i] = fnType.In(i)
	}

	var outputType reflect.Type
	if fnType.NumOut() == resultAndErrorOutput {
		outputType = fnType.Out(0)
	}

	return &functionInfo{
		inputTypes: inputTypes,
		outputType: outputType,
	}, nil
}

func validateFunction(fnType reflect.Type) error {
	if fnType.Kind() != reflect.Func {
		return errors.ErrMustBeFunction
	}

	if err := validateParameters(fnType); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if err := validateReturns(fnType); err != nil {
		return fmt.Errorf("invalid returns: %w", err)
	}

	return nil
}

func validateParameters(fnType reflect.Type) error {
	// Must have at least context parameter
	if fnType.NumIn() < 1 {
		return errors.ErrMustHaveAtLeastContext
	}

	// Check if variadic (not supported per spec)
	if fnType.IsVariadic() {
		return errors.ErrVariadicNotSupported
	}

	// First parameter must be context.Context
	firstParam := fnType.In(0)
	if !firstParam.Implements(contextInterface) {
		return errors.ErrFirstParamMustBeContext
	}
	return nil
}

func validateReturns(fnType reflect.Type) error {
	switch fnType.NumOut() {
	case noOutput:
		return errors.ErrMustReturnAtLeastError
	case onlyErrorOutput:
		if !fnType.Out(0).Implements(errorInterface) {
			return errors.ErrSingleReturnMustBeError
		}
	case resultAndErrorOutput:
		if !fnType.Out(1).Implements(errorInterface) {
			return errors.ErrSecondReturnMustBeError
		}
	default:
		return errors.ErrTooManyReturnValues
	}
	return nil
}
