package lyra

import (
	"context"
	"reflect"

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
		_ = i
		params = append(params, reflect.ValueOf(value))
	}
	_ = types

	return params, nil
}
