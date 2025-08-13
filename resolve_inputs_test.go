package lyra

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

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
