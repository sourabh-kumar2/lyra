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
	params := []reflect.Value{reflect.ValueOf(ctx)}
	specs, types := task.GetInputParams()

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
			value, err = fieldTypeFromPath(value, spec.Field)
			if err != nil {
				return nil, errors.Wrapf(err, "parameter %d", i+2)
			}
		}

		if reflect.TypeOf(value) != types[i+1] { // first is context.Context
			return nil, errors.Wrapf(
				errors.ErrInvalidParamType,
				"parameter %d -> exptected type %s, got %s",
				i+2, // array offset (1) + first param is context (1) = 2
				types[i+1],
				reflect.TypeOf(value),
			)
		}
		params = append(params, reflect.ValueOf(value))
	}

	return params, nil
}

//nolint:err113 // static error because its too specific
//revive:disable-next-line:cognitive-complexity // struct walking algo is complex.
func fieldTypeFromPath(value any, path string) (any, error) {
	if value == nil {
		return nil, stderr.New("value is nil")
	}

	v := reflect.ValueOf(value)
	parts := strings.Split(path, ".")

	for _, part := range parts {
		// Dereference pointers
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil, fmt.Errorf("nil pointer while accessing %q", part)
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %q is not a struct (found %s)", part, v.Kind())
		}

		f := v.FieldByName(part)
		if !f.IsValid() {
			return nil, fmt.Errorf("field %q not found in %s", part, v.Type())
		}

		v = f
	}

	return v.Interface(), nil
}
