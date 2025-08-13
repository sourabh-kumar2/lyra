package lyra

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/errors"
	"github.com/sourabh-kumar2/lyra/internal"
)

func TestLyraResolveInputsContextOnly(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask(
		"testTask",
		func(ctx context.Context) (string, error) { return "test", nil },
		[]internal.InputSpec{},
	)

	require.NoError(t, err)

	results := NewResult()
	ctx := context.Background()

	args, err := resolveInputs(ctx, task, results)

	require.NoError(t, err)
	require.Len(t, args, 1) // Just context

	// Verify context type
	ctxArg := args[0].Interface()
	require.Implements(t, (*context.Context)(nil), ctxArg)
}

func TestLyraResolveInputsRuntimeInput(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask("testTask",
		func(ctx context.Context, userID int) (string, error) { return "test", nil },
		[]internal.InputSpec{UseRun("userID")})
	require.NoError(t, err)

	results := NewResult()
	results.set("userID", 123) // Simulate runtime input

	args, err := resolveInputs(context.Background(), task, results)

	require.NoError(t, err)
	require.Len(t, args, 2) // context + userID

	// Verify userID value
	userIDArg := args[1].Interface()
	require.Equal(t, 123, userIDArg)
}

func TestLyraResolveInputsTaskResultInput(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask("consumer",
		func(ctx context.Context, user string) (string, error) { return "processed", nil },
		[]internal.InputSpec{Use("producer")})
	require.NoError(t, err)

	results := NewResult()
	results.set("producer", "user_data") // Simulate previous task result

	args, err := resolveInputs(context.Background(), task, results)

	require.NoError(t, err)
	require.Len(t, args, 2)
	require.Equal(t, "user_data", args[1].Interface())
}

func TestLyraResolveInputsMultipleInputs(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask("multiTask",
		func(ctx context.Context, userID int, userData string, isActive bool) (string, error) {
			return "result", nil
		},
		[]internal.InputSpec{
			UseRun("userID"),
			Use("fetchUser"),
			UseRun("active"),
		})
	require.NoError(t, err)

	results := NewResult()
	results.set("userID", 456)
	results.set("fetchUser", "john_doe")
	results.set("active", true)

	args, err := resolveInputs(context.Background(), task, results)

	require.NoError(t, err)
	require.Len(t, args, 4) // context + 3 params
	require.Equal(t, 456, args[1].Interface())
	require.Equal(t, "john_doe", args[2].Interface())
	require.Equal(t, true, args[3].Interface())
}

func TestLyraResolveInputsTypeMismatch(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask("typeMismatch",
		func(ctx context.Context, userID int) (string, error) { return "test", nil },
		[]internal.InputSpec{UseRun("userID")})
	require.NoError(t, err)

	results := NewResult()
	results.set("userID", "string_instead_of_int") // Wrong type

	args, err := resolveInputs(context.Background(), task, results)

	require.ErrorIs(t, err, errors.ErrInvalidParamType)
	require.Nil(t, args)
}

func TestLyraResolveInputsMissingRuntimeInput(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask("missingInput",
		func(ctx context.Context, userID int) (string, error) { return "test", nil },
		[]internal.InputSpec{UseRun("userID")})
	require.NoError(t, err)

	results := NewResult() // Empty results

	args, err := resolveInputs(context.Background(), task, results)

	require.ErrorIs(t, err, errors.ErrTaskNotFound)
	require.Nil(t, args)
	require.Contains(t, err.Error(), "userID")
}

func TestLyraResolveInputsMissingTaskResult(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask("missingDep",
		func(ctx context.Context, userData string) (string, error) { return "test", nil },
		[]internal.InputSpec{Use("nonExistentTask")})
	require.NoError(t, err)

	results := NewResult()

	args, err := resolveInputs(context.Background(), task, results)

	require.ErrorIs(t, err, errors.ErrTaskNotFound)
	require.Nil(t, args)
	require.Contains(t, err.Error(), "nonExistentTask")
}

func TestLyraResolveInputsNilValue(t *testing.T) {
	t.Parallel()

	task, err := internal.NewTask("nilValue",
		func(ctx context.Context, user *string) (string, error) { return "test", nil },
		[]internal.InputSpec{Use("producer")})
	require.NoError(t, err)

	results := NewResult()
	results.set("producer", (*string)(nil)) // Nil pointer

	args, err := resolveInputs(context.Background(), task, results)

	require.NoError(t, err)
	require.Len(t, args, 2)
	require.True(t, args[1].IsNil())
}
