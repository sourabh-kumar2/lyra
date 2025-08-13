package lyra

import (
	"context"
	stderr "errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sourabh-kumar2/lyra/errors"
	"github.com/sourabh-kumar2/lyra/internal"
)

func resolveInputs(
	ctx context.Context,
	task *internal.Task,
	results *Result,
) ([]reflect.Value, error) {
	specs, types := task.GetInputParams()
	args := make([]reflect.Value, len(types))
	args[0] = reflect.ValueOf(ctx) // First arg is always context

	for i, spec := range specs {
		value, err := results.Get(spec.Source)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to get %v for task %q, did you miss to set in run config",
				spec.Source,
				task.GetID(),
			)
		}
		if spec.Field != "" {
			value, err = extractNestedField(value, spec.Field)
			if err != nil {
				return nil, errors.Wrapf(err, "parameter %d", i+2)
			}
		}

		expectedType := types[i+1] // +1 to skip context
		actualValue := reflect.ValueOf(value)
		if !actualValue.Type().AssignableTo(expectedType) {
			return nil, errors.Wrapf(
				errors.ErrInvalidParamType,
				"parameter %d -> exptected type %s, got %s",
				i+2, // array offset (1) + first param is context (1) = 2
				expectedType,
				actualValue.Type(),
			)
		}
		args[i+1] = actualValue
	}

	return args, nil
}

//nolint:err113 // static error because its too specific
//revive:disable-next-line:cognitive-complexity // struct walking algo is complex.
func extractNestedField(value any, path string) (any, error) {
	if value == nil {
		return nil, stderr.New("value is nil")
	}

	current := reflect.ValueOf(value)
	fields := strings.Split(path, ".")

	for _, fieldName := range fields {
		if current.Kind() == reflect.Ptr && current.IsNil() {
			return nil, fmt.Errorf("nil pointer encountered while accessing field %q", fieldName)
		}

		// Dereference pointers
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %q is not a struct (found %s)", fieldName, current.Kind())
		}

		fieldValue := current.FieldByName(fieldName)
		if !fieldValue.IsValid() {
			return nil, fmt.Errorf("field %q not found in type %v", fieldName, current.Type())
		}

		if !fieldValue.CanInterface() {
			return nil, fmt.Errorf("field %q is not exported in type %v", fieldName, current.Type())
		}

		current = fieldValue
	}

	return current.Interface(), nil
}
