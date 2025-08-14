package lyra

import (
	"context"
	stderr "errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourabh-kumar2/lyra/errors"
	"github.com/sourabh-kumar2/lyra/internal"
)

//revive:disable:use-errors-new,unnecessary-format

func TestNew(t *testing.T) {
	t.Parallel()

	l := New()
	require.NotNil(t, l)
	require.NotNil(t, l.tasks)
	require.Len(t, l.tasks, 0)
	require.Nil(t, l.error)
}

func TestDo(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name              string
		tasks             []testTask
		expectedTaskCount int
		expectedErr       error
	}{
		{
			name: "single task",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
			},
			expectedTaskCount: 1,
		},
		{
			name: "multiple tasks",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
				{
					id:         "task-2",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
				{
					id:         "task-3",
					fn:         anotherValidTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
				{
					id:         "task-4",
					fn:         dependentTask,
					inputSpecs: []internal.InputSpec{Use("task-2")},
				},
			},
			expectedTaskCount: 4,
		},
		{
			name: "valid task function with empty inputSpecs",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: nil,
				},
			},
			expectedTaskCount: 1,
		},
		{
			name: "invalid task function with empty inputSpecs",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         invalidTask,
					inputSpecs: nil,
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrMustHaveAtLeastContext,
		},
		{
			name: "empty id",
			tasks: []testTask{
				{
					id:         "",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskIDCannotBeEmpty,
		},
		{
			name: "empty id with whitespace",
			tasks: []testTask{
				{
					id:         "     ",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskIDCannotBeEmpty,
		},
		{
			name: "duplicate task id",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         anotherValidTask,
					inputSpecs: []internal.InputSpec{UseRun("orderID")},
				},
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: nil,
				},
			},
			expectedTaskCount: 1,
			expectedErr:       errors.ErrDuplicateTask,
		},
		{
			name: "invalid task function",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         invalidTask,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrMustHaveAtLeastContext,
		},
		{
			name: "wrong input count less than expected",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskParamCountMismatch,
		},
		{
			name: "wrong input count greater than expected",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrTaskParamCountMismatch,
		},
		{
			name: "nil function",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         nil,
					inputSpecs: []internal.InputSpec{},
				},
			},
			expectedTaskCount: 0,
			expectedErr:       errors.ErrMustBeFunction,
		},
		{
			name: "adding tasks even after error",
			tasks: []testTask{
				{
					id:         "task-1",
					fn:         validTaskWithNoInput,
					inputSpecs: []internal.InputSpec{},
				},
				{
					id:         "task-2",
					fn:         validTask,
					inputSpecs: []internal.InputSpec{UseRun("userID")},
				},
				{
					id:         "task-3",
					fn:         anotherValidTask,
					inputSpecs: []internal.InputSpec{UseRun("orderID"), UseRun("userID")},
				},
				{
					id:         "task-4",
					fn:         dependentTask,
					inputSpecs: []internal.InputSpec{Use("task-2")},
				},
			},
			expectedTaskCount: 3,
			expectedErr:       errors.ErrTaskParamCountMismatch,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := New()
			for _, task := range tc.tasks {
				l.Do(task.id, task.fn, task.inputSpecs...)
			}
			require.ErrorIs(t, l.error, tc.expectedErr)
			require.Len(t, l.tasks, tc.expectedTaskCount)
		})
	}
}

func TestDoConcurrency(t *testing.T) {
	t.Parallel()

	l := New()

	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			l.Do(fmt.Sprintf("task-%d", id), validTaskWithNoInput)
		}(i)
	}

	wg.Wait()
	require.Len(t, l.tasks, 10)
}

func TestRunBuildError(t *testing.T) {
	t.Parallel()

	result, err := New().
		Do("task-1", invalidTask).
		Run(context.Background(), nil)

	require.ErrorIs(t, err, errors.ErrMustHaveAtLeastContext)
	require.Nil(t, result)
}

func TestRunEmptyDAG(t *testing.T) {
	t.Parallel()

	l := New()
	result, err := l.Run(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestRunEmptyDAGWithRunInputs(t *testing.T) {
	t.Parallel()

	runInputs := map[string]any{
		"userID":  123,
		"orderID": 456,
	}

	l := New()
	result, err := l.Run(context.Background(), runInputs)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, result.data, runInputs)
}

func TestRunSingleTaskNoInputs(t *testing.T) {
	t.Parallel()

	l := New().
		Do("hello", func(ctx context.Context) (string, error) {
			return "world", nil
		})

	result, err := l.Run(context.Background(), nil)

	require.NoError(t, err)
	helloResult, err := result.Get("hello")
	require.NoError(t, err)
	require.Equal(t, "world", helloResult)
}

func TestRunSingleTaskWithRuntimeInput(t *testing.T) {
	t.Parallel()

	l := New().
		Do("greet", func(ctx context.Context, name string) (string, error) {
			return fmt.Sprintf("Hello %s", name), nil
		}, UseRun("name"))

	result, err := l.Run(context.Background(), map[string]any{"name": "Alice"})

	require.NoError(t, err)
	greeting, err := result.Get("greet")
	require.NoError(t, err)
	require.Equal(t, "Hello Alice", greeting)
}

func TestRunLinearDependency(t *testing.T) {
	t.Parallel()

	l := New().
		Do("task1", func(ctx context.Context, userID int) (string, error) {
			return fmt.Sprintf("user_%d", userID), nil
		}, UseRun("userID")).
		Do("task2", func(ctx context.Context, user string) (string, error) {
			return user + "_processed", nil
		}, Use("task1"))

	result, err := l.Run(context.Background(), map[string]any{"userID": 123})

	require.NoError(t, err)

	task1Result, err := result.Get("task1")
	require.NoError(t, err)
	require.Equal(t, "user_123", task1Result)

	task2Result, err := result.Get("task2")
	require.NoError(t, err)
	require.Equal(t, "user_123_processed", task2Result)
}

func TestRunDeepLinearChain(t *testing.T) {
	t.Parallel()

	l := New().
		Do("step1", func(ctx context.Context, input int) (int, error) {
			return input * 2, nil
		}, UseRun("input")).
		Do("step2", func(ctx context.Context, val int) (int, error) {
			return val + 10, nil
		}, Use("step1")).
		Do("step3", func(ctx context.Context, val int) (int, error) {
			return val * 3, nil
		}, Use("step2")).
		Do("step4", func(ctx context.Context, val int) (string, error) {
			return fmt.Sprintf("result_%d", val), nil
		}, Use("step3"))

	result, err := l.Run(context.Background(), map[string]any{"input": 5})

	require.NoError(t, err)

	finalResult, err := result.Get("step4")
	require.NoError(t, err)
	// 5 * 2 = 10, 10 + 10 = 20, 20 * 3 = 60
	require.Equal(t, "result_60", finalResult)
}

//nolint:err113 // test case.
func TestRunDiamondDependency(t *testing.T) {
	t.Parallel()

	l := New().
		Do("fetchUser", func(ctx context.Context, userID int) (User, error) {
			return User{ID: userID, Name: "Alice", Email: "alice@example.com"}, nil
		}, UseRun("userID")).
		Do("validateEmail", func(ctx context.Context, user User) (bool, error) {
			return (user.Email) != "", nil
		}, Use("fetchUser")).
		Do("formatName", func(ctx context.Context, user User) (string, error) {
			return fmt.Sprintf("Mr/Ms %s", user.Name), nil
		}, Use("fetchUser")).
		Do("createProfile", func(ctx context.Context, user User, isValid bool, formatted string) (string, error) {
			if !isValid {
				return "", fmt.Errorf("invalid email")
			}
			return fmt.Sprintf("Profile: %s (%s)", formatted, user.Email), nil
		}, Use("fetchUser"), Use("validateEmail"), Use("formatName"))

	result, err := l.Run(context.Background(), map[string]any{"userID": 123})

	require.NoError(t, err)

	profile, err := result.Get("createProfile")
	require.NoError(t, err)
	require.Equal(t, "Profile: Mr/Ms Alice (alice@example.com)", profile)
}

func TestRunComplexWorkflow(t *testing.T) {
	t.Parallel()

	l := New().
		Do("fetchUser", func(ctx context.Context, userID int) (User, error) {
			return User{
				ID:      userID,
				Name:    "Bob",
				Address: Address{Street: "123 Main St", City: "Boston"},
			}, nil
		}, UseRun("userID")).
		Do("fetchOrders", func(ctx context.Context, userID int) ([]Order, error) {
			return []Order{
				{ID: 1, UserID: userID, Amount: 100.0},
				{ID: 2, UserID: userID, Amount: 250.0},
			}, nil
		}, UseRun("userID")).
		Do("calculateTotal", func(ctx context.Context, orders []Order) (float64, error) {
			total := 0.0
			for _, order := range orders {
				total += order.Amount
			}
			return total, nil
		}, Use("fetchOrders")).
		Do("generateReport", func(ctx context.Context, user User, orders []Order, total float64) (Report, error) {
			return Report{
				UserName:   user.Name,
				OrderCount: len(orders),
				TotalSpent: total,
			}, nil
		}, Use("fetchUser"), Use("fetchOrders"), Use("calculateTotal"))

	result, err := l.Run(context.Background(), map[string]any{"userID": 456})

	require.NoError(t, err)

	report, err := result.Get("generateReport")
	require.NoError(t, err)

	expectedReport := Report{
		UserName:   "Bob",
		OrderCount: 2,
		TotalSpent: 350.0,
	}
	require.Equal(t, expectedReport, report)
}

func TestRunLargeFanInFanOut(t *testing.T) {
	t.Parallel()

	l := New().
		Do("source", func(ctx context.Context, input int) (int, error) {
			return input, nil
		}, UseRun("input"))

	// Fan out: create 5 parallel tasks
	for i := 1; i <= 5; i++ {
		taskID := fmt.Sprintf("process_%d", i)
		l.Do(taskID, func(ctx context.Context, val int) (int, error) {
			return val * i, nil
		}, Use("source"))
	}

	// Fan in: aggregate all results
	l.Do("aggregate", func(ctx context.Context, r1, r2, r3, r4, r5 int) (int, error) {
		return r1 + r2 + r3 + r4 + r5, nil
	}, Use("process_1"), Use("process_2"), Use("process_3"),
		Use("process_4"), Use("process_5"))

	result, err := l.Run(context.Background(), map[string]any{"input": 10})

	require.NoError(t, err)

	aggregate, err := result.Get("aggregate")
	require.NoError(t, err)
	// 10*1 + 10*2 + 10*3 + 10*4 + 10*5 = 10 + 20 + 30 + 40 + 50 = 150
	require.Equal(t, 150, aggregate)
}

func TestRunNestedFieldAccess(t *testing.T) {
	t.Parallel()

	l := New().
		Do("getUser", func(ctx context.Context) (User, error) {
			return User{
				Name: "Charlie",
				Address: Address{
					Street: "456 Oak Ave",
					City:   "Chicago",
				},
			}, nil
		}).
		Do("processCity", func(ctx context.Context, city string) (string, error) {
			return fmt.Sprintf("Processing in %s", city), nil
		}, Use("getUser", "Address", "City"))

	result, err := l.Run(context.Background(), nil)

	require.NoError(t, err)

	processed, err := result.Get("processCity")
	require.NoError(t, err)
	require.Equal(t, "Processing in Chicago", processed)
}

//nolint:err113 // test case.
func TestRunTaskExecutionError(t *testing.T) {
	t.Parallel()
	l := New().
		Do("failingTask", func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("something went wrong")
		}).
		Do("dependentTask", func(ctx context.Context, input string) (string, error) {
			return "processed_" + input, nil
		}, Use("failingTask"))

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "something went wrong")
}

//nolint:err113 // test case.
func TestRunTaskExecutionErrorOnlyErrorOutput(t *testing.T) {
	t.Parallel()
	l := New().
		Do("failingTask", func(ctx context.Context) error {
			return fmt.Errorf("something went wrong")
		}).
		Do("dependentTask", func(ctx context.Context, input string) (string, error) {
			return "processed_" + input, nil
		}, Use("failingTask"))

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "something went wrong")
}

func TestRunMissingRuntimeInput(t *testing.T) {
	t.Parallel()
	l := New().
		Do("needsInput", func(ctx context.Context, name string) (string, error) {
			return "Hello " + name, nil
		}, UseRun("name"))

	result, err := l.Run(context.Background(), map[string]any{"wrongKey": "value"})

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "name")
}

func TestRunTypeMismatch(t *testing.T) {
	t.Parallel()

	l := New().
		Do("producer", func(ctx context.Context) (string, error) {
			return "text_result", nil
		}).
		Do("consumer", func(ctx context.Context, num int) (string, error) {
			return fmt.Sprintf("number_%d", num), nil
		}, Use("producer")) // string -> int mismatch

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "type")
}

func TestRunCyclicDependency(t *testing.T) {
	t.Parallel()
	l := New().
		Do("taskA", func(ctx context.Context, input string) (string, error) {
			return "A_" + input, nil
		}, Use("taskB")).
		Do("taskB", func(ctx context.Context, input string) (string, error) {
			return "B_" + input, nil
		}, Use("taskA"))

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "cyclic")
}

func TestRunMissingDependency(t *testing.T) {
	t.Parallel()

	l := New().
		Do("task", func(ctx context.Context, input string) (string, error) {
			return "processed_" + input, nil
		}, Use("nonExistentTask"))

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "nonExistentTask")
}

func TestRunNilTaskResult(t *testing.T) {
	t.Parallel()

	l := New().
		Do("nilTask", func(ctx context.Context) (*string, error) {
			//nolint:nilnil // testing specific scenario
			return nil, nil
		}).
		Do("consumer", func(ctx context.Context, input *string) (string, error) {
			if input == nil {
				return "got_nil", nil
			}
			return *input, nil
		}, Use("nilTask"))

	result, err := l.Run(context.Background(), nil)

	require.NoError(t, err)

	consumerResult, err := result.Get("consumer")
	require.NoError(t, err)
	require.Equal(t, "got_nil", consumerResult)
}

func TestRunContextTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	l := New().
		Do("slowTask", func(ctx context.Context) (string, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return "completed", nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		})

	result, err := l.Run(ctx, nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "context")
}

func TestRunContextPropagation(t *testing.T) {
	t.Parallel()

	type contextKey string
	const testKey contextKey = "testValue"

	ctx := context.WithValue(context.Background(), testKey, "propagated")

	l := New().
		Do("checkContext", func(ctx context.Context) (string, error) {
			if val := ctx.Value(testKey); val != nil {
				s, ok := val.(string)
				if !ok {
					return "", nil
				}
				return s, nil
			}
			return "not_found", nil
		})

	result, err := l.Run(ctx, nil)

	require.NoError(t, err)

	ctxResult, err := result.Get("checkContext")
	require.NoError(t, err)
	require.Equal(t, "propagated", ctxResult)
}

func TestDoDuplicateTaskID(t *testing.T) {
	t.Parallel()

	l := New().
		Do("task", func(ctx context.Context) (string, error) {
			return "first", nil
		}).
		Do("task", func(ctx context.Context) (string, error) {
			return "second", nil
		})

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "duplicate")
}

func TestDoInvalidFunctionSignature(t *testing.T) {
	t.Parallel()

	l := New().
		Do("invalidTask", func() string { // Missing context, missing error return
			return "invalid"
		})

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
}

func TestRunLargeDAGPerformance(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	l := New()

	// Create a large DAG: 100 tasks with complex dependencies
	l.Do("root", func(ctx context.Context) (int, error) {
		return 1, nil
	})

	// Create 10 levels with 10 tasks each
	for level := 1; level <= 10; level++ {
		for task := 1; task <= 10; task++ {
			taskID := fmt.Sprintf("L%d_T%d", level, task)
			prevTaskID := "root"
			if level > 1 {
				prevTaskID = fmt.Sprintf("L%d_T%d", level-1, task)
			}

			l.Do(taskID, func(ctx context.Context, input int) (int, error) {
				return input + 1, nil
			}, Use(prevTaskID))
		}
	}

	start := time.Now()
	result, err := l.Run(context.Background(), nil)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Should complete within reasonable time
	require.Less(t, duration, 5*time.Second, "DAG execution took too long: %v", duration)

	t.Logf("Large DAG (100 tasks) executed in %v", duration)
}

func TestConcurrentStageExecution(t *testing.T) {
	t.Parallel()

	// Track execution order to verify concurrency
	var mu sync.Mutex
	var executionOrder []string
	var executionTimes []time.Time

	recordExecution := func(taskID string) {
		mu.Lock()
		defer mu.Unlock()
		executionOrder = append(executionOrder, taskID)
		executionTimes = append(executionTimes, time.Now())
	}

	l := New().
		Do("source", func(ctx context.Context) (int, error) {
			recordExecution("source")
			return 10, nil
		}).
		// These should execute concurrently (same stage)
		Do("parallel1", func(ctx context.Context, val int) (int, error) {
			recordExecution("parallel1")
			time.Sleep(50 * time.Millisecond) // Simulate work
			return val * 2, nil
		}, Use("source")).
		Do("parallel2", func(ctx context.Context, val int) (int, error) {
			recordExecution("parallel2")
			time.Sleep(50 * time.Millisecond) // Simulate work
			return val * 3, nil
		}, Use("source")).
		Do("parallel3", func(ctx context.Context, val int) (int, error) {
			recordExecution("parallel3")
			time.Sleep(50 * time.Millisecond) // Simulate work
			return val * 4, nil
		}, Use("source"))

	start := time.Now()
	result, err := l.Run(context.Background(), nil)
	totalDuration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify results
	p1, err := result.Get("parallel1")
	require.NoError(t, err)
	require.Equal(t, 20, p1)

	p2, err := result.Get("parallel2")
	require.NoError(t, err)
	require.Equal(t, 30, p2)

	p3, err := result.Get("parallel3")
	require.NoError(t, err)
	require.Equal(t, 40, p3)

	// Verify concurrency: should complete in ~50ms, not 150ms
	require.Less(t, totalDuration, 120*time.Millisecond,
		"Tasks should execute concurrently, took %v", totalDuration)

	// Verify execution order: source should be first
	require.Equal(t, "source", executionOrder[0])

	// The parallel tasks should start within a short time window
	parallelStartTimes := executionTimes[1:] // Skip source
	maxTimeDiff := time.Duration(0)
	for i := 1; i < len(parallelStartTimes); i++ {
		diff := parallelStartTimes[i].Sub(parallelStartTimes[0])
		if diff > maxTimeDiff {
			maxTimeDiff = diff
		}
	}

	// Parallel tasks should start within 10ms of each other
	require.Less(t, maxTimeDiff, 10*time.Millisecond,
		"Parallel tasks should start nearly simultaneously")

	t.Logf("Concurrent execution completed in %v", totalDuration)
	t.Logf("Execution order: %v", executionOrder)
}

// Test error handling in concurrent execution.
func TestConcurrentStageWithError(t *testing.T) {
	t.Parallel()

	l := New().
		Do("source", func(ctx context.Context) (int, error) {
			return 10, nil
		}).
		Do("success", func(ctx context.Context, val int) (int, error) {
			time.Sleep(100 * time.Millisecond) // Longer than failure
			return val * 2, nil
		}, Use("source")).
		Do("failure", func(ctx context.Context, val int) (int, error) {
			time.Sleep(20 * time.Millisecond)
			//nolint:err113 // it's a test error
			return 0, stderr.New("task failed")
		}, Use("source"))

	result, err := l.Run(context.Background(), nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "task failed")
}

// Test context cancellation in concurrent execution.
func TestConcurrentStageContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	l := New().
		Do("source", func(ctx context.Context) (int, error) {
			return 10, nil
		}).
		Do("long1", func(ctx context.Context, val int) (int, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return val, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		}, Use("source")).
		Do("long2", func(ctx context.Context, val int) (int, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return val, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		}, Use("source"))

	// Cancel context after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := l.Run(ctx, nil)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "context")
}

type User struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Address Address `json:"address"`
}

type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
	State  string `json:"state"`
}

type Order struct {
	ID     int     `json:"id"`
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
	Items  []Item  `json:"items"`
}

type Item struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type Report struct {
	UserID     int     `json:"user_id"`
	UserName   string  `json:"user_name"`
	TotalSpent float64 `json:"total_spent"`
	OrderCount int     `json:"order_count"`
}

func validTask(ctx context.Context, userID string) (User, error) {
	return User{}, nil
}

func validTaskWithNoInput(ctx context.Context) error {
	return nil
}

func anotherValidTask(ctx context.Context, orderID string) error {
	return nil
}

//nolint:gocritic // This is test case.
func dependentTask(ctx context.Context, user User) error {
	return nil
}

func invalidTask() {
}

type testTask struct {
	id         string
	fn         any
	inputSpecs []internal.InputSpec
}
