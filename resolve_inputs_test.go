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
