package lyra

import (
	"context"
	"reflect"

	"github.com/sourabh-kumar2/lyra/internal"
)

func resolveInputs(
	ctx context.Context,
	task *internal.Task,
	results *Result,
) ([]reflect.Value, error) {
	params := []reflect.Value{reflect.ValueOf(ctx)}

	_ = task
	_ = results
	return params, nil
}
