package internal

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/errors"
)

// User defines a test user.
type User struct {
	ID    string
	Name  string
	Email string
}

// Order defines a test order.
type Order struct {
	ID     string
	Amount float64
}

// Report defines a test report.
type Report struct {
	UserID string
	Total  float64
}

// CustomError error type for testing.
type CustomError struct {
	message string
}

// Error returns string error.
func (e CustomError) Error() string {
	return e.message
}

//nolint:nilnil // this is a test function
func TestAnalyzeFunctionSignatureValidFunctions(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name           string
		fn             any
		expectedInputs int          // Number of inputs (excluding context)
		expectedOutput reflect.Type // Expected output type (nil if only error)
	}{
		{
			name: "no_input_only_error",
			fn: func(ctx context.Context) error {
				return nil
			},
			expectedInputs: 0,
			expectedOutput: nil,
		},
		{
			name: "no_input_with_result",
			fn: func(context.Context) (User, error) {
				return User{}, nil
			},
			expectedInputs: 0,
			expectedOutput: reflect.TypeOf(User{}),
		},
		{
			name: "single_input_with_result",
			fn: func(ctx context.Context, userID string) (User, error) {
				return User{}, nil
			},
			expectedInputs: 1,
			expectedOutput: reflect.TypeOf(User{}),
		},
		{
			name: "multiple_inputs_with_result",
			fn: func(ctx context.Context, user User, orders []Order) (Report, error) {
				return Report{}, nil
			},
			expectedInputs: 2,
			expectedOutput: reflect.TypeOf(Report{}),
		},
		{
			name: "primitive_types",
			fn: func(ctx context.Context, id int, name string, active bool) (string, error) {
				return "", nil
			},
			expectedInputs: 3,
			expectedOutput: reflect.TypeOf(""),
		},
		{
			name: "interface_input",
			fn: func(ctx context.Context, data any) ([]byte, error) {
				return nil, nil
			},
			expectedInputs: 1,
			expectedOutput: reflect.TypeOf([]byte{}),
		},
		{
			name: "pointer_types",
			fn: func(ctx context.Context, user *User) (*Report, error) {
				return nil, nil
			},
			expectedInputs: 1,
			expectedOutput: reflect.TypeOf(&Report{}),
		},
		{
			name: "slice_types",
			fn: func(ctx context.Context, users []User, ids []string) ([]Report, error) {
				return nil, nil
			},
			expectedInputs: 2,
			expectedOutput: reflect.TypeOf([]Report{}),
		},
		{
			name: "map_types",
			fn: func(ctx context.Context, data map[string]any) (map[string]User, error) {
				return nil, nil
			},
			expectedInputs: 1,
			expectedOutput: reflect.TypeOf(map[string]User{}),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			fnInfo, err := analyzeFunctionSignature(tc.fn)
			require.NoError(t, err, "expected valid function")

			// Check input count (excluding context)
			actualInputs := len(fnInfo.inputTypes) - 1 // Subtract 1 for context
			require.Equal(t, tc.expectedInputs, actualInputs)

			// Check output type
			if tc.expectedOutput == nil {
				require.Nil(t, fnInfo.outputType, "expected no output type")
			} else {
				require.Equal(t, tc.expectedOutput, fnInfo.outputType)
			}

			// Verify first parameter is context.Context
			if len(fnInfo.inputTypes) > 0 {
				contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
				require.Truef(
					t,
					fnInfo.inputTypes[0].Implements(contextType),
					"first parameter should implement context.Context, got %v",
					fnInfo.inputTypes[0],
				)
			}
		})
	}
}

func TestAnalyzeFunctionSignatureInvalidFunctions(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name        string
		fn          any
		expectedErr error
	}{
		{
			name:        "not_a_function",
			fn:          "not a function",
			expectedErr: errors.ErrMustBeFunction,
		},
		{
			name:        "nil_input",
			fn:          nil,
			expectedErr: errors.ErrMustBeFunction,
		},
		{
			name: "no_parameters",
			fn: func() error {
				return nil
			},
			expectedErr: errors.ErrMustHaveAtLeastContext,
		},
		{
			name: "first_param_not_context",
			fn: func(userID string) error {
				return nil
			},
			expectedErr: errors.ErrFirstParamMustBeContext,
		},
		{
			name: "no_return_values",
			fn: func(ctx context.Context) {
				// No return
			},
			expectedErr: errors.ErrMustReturnAtLeastError,
		},
		{
			name: "too_many_returns",
			fn: func(ctx context.Context) (string, int, error) {
				return "", 0, nil
			},
			expectedErr: errors.ErrTooManyReturnValues,
		},
		{
			name: "single_return_not_error",
			fn: func(ctx context.Context) string {
				return ""
			},
			expectedErr: errors.ErrSingleReturnMustBeError,
		},
		{
			name: "second_return_not_error",
			fn: func(ctx context.Context) (string, int) {
				return "", 0
			},
			expectedErr: errors.ErrSecondReturnMustBeError,
		},
		{
			name: "custom_context_type",
			fn: func(ctx *User) error { // Wrong context type
				return nil
			},
			expectedErr: errors.ErrFirstParamMustBeContext,
		},
		{
			name: "variadic_function",
			fn: func(ctx context.Context, ids ...string) error {
				return nil
			},
			expectedErr: errors.ErrVariadicNotSupported,
		},
		{
			name: "variadic_function",
			fn: func(ids ...string) error {
				return nil
			},
			expectedErr: errors.ErrVariadicNotSupported,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := analyzeFunctionSignature(tc.fn)
			require.Error(t, err, "expected error for invalid function, got valid result")
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestAnalyzeFunctionSignatureEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("custom_error_implementation", func(t *testing.T) {
		fn := func(ctx context.Context) CustomError {
			return CustomError{}
		}

		fnInfo, err := analyzeFunctionSignature(fn)
		require.NoError(t, err, "custom error type should be valid")

		require.Nil(
			t,
			fnInfo.outputType,
			"custom error return should be treated as error-only",
		)
	})

	t.Run("interface_that_implements_error", func(t *testing.T) {
		fn := func(ctx context.Context, data string) (any, error) {
			return 5, nil
		}

		fnInfo, err := analyzeFunctionSignature(fn)
		require.NoError(t, err, "interface{} output should be valid")

		expectedType := reflect.TypeOf((*any)(nil)).Elem()
		require.Equal(t, expectedType, fnInfo.outputType)
	})

	t.Run("context_derived_types", func(t *testing.T) {
		// Test with context.WithCancel derived context
		fn := func(ctx context.Context) error {
			return nil
		}

		fnInfo, err := analyzeFunctionSignature(fn)
		require.NoError(t, err, "context.Context should be valid")
		require.Len(t, fnInfo.inputTypes, 1)
	})
}

func TestFunctionInfoStructure(t *testing.T) {
	t.Parallel()

	fn := func(ctx context.Context, user User, count int) (Report, error) {
		return Report{}, nil
	}

	fnInfo, err := analyzeFunctionSignature(fn)

	require.NoError(t, err)
	require.Len(t, fnInfo.inputTypes, 3)
	require.True(
		t,
		fnInfo.inputTypes[0].Implements(contextInterface),
		"input[0]: expected type implementing context",
	)
	require.Equal(
		t,
		reflect.TypeOf(User{}),
		fnInfo.inputTypes[1],
		"input[1]: type mismatch",
	)
	require.Equal(t, reflect.TypeOf(0), fnInfo.inputTypes[2], "input[2]: type mismatch")

	expectedOutput := reflect.TypeOf(Report{})
	require.Equal(t, expectedOutput, fnInfo.outputType)
}

// Benchmark tests for performance analysis.
func BenchmarkAnalyzeFunctionSignatureSimple(b *testing.B) {
	fn := func(ctx context.Context, id string) (User, error) {
		return User{}, nil
	}

	b.ResetTimer()
	for range b.N {
		_, err := analyzeFunctionSignature(fn)
		if err != nil {
			b.Fatal(err)
		}
	}
}

//nolint:nilnil // Running benchmark
func BenchmarkAnalyzeFunctionSignatureComplex(b *testing.B) {
	// revive:disable-next-line:line-length-limit // This is function signature
	fn := func(ctx context.Context, user User, orders []Order, metadata map[string]any) (*Report, error) {
		return nil, nil
	}

	b.ResetTimer()
	for range b.N {
		_, err := analyzeFunctionSignature(fn)
		if err != nil {
			b.Fatal(err)
		}
	}
}
