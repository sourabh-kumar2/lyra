package lyra

import (
	"context"
	"fmt"
	"github.com/sourabh-kumar2/lyra/internal"
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkSimpleDAG tests a basic 3-task DAG.
func BenchmarkSimpleDAG(b *testing.B) {
	for range b.N {
		l := New()
		l.Do("fetchUser", fetchUser, UseRun("userID"))
		l.Do("fetchOrders", fetchOrders, UseRun("userID"))
		l.Do("generateReport", generateReport, Use("fetchUser"), Use("fetchOrders"))

		_, err := l.Run(context.Background(), map[string]any{"userID": 123})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComplexDAG tests a more complex DAG with parallel branches.
func BenchmarkComplexDAG(b *testing.B) {
	for range b.N {
		l := New()
		l.Do("fetchUser", fetchUser, UseRun("userID"))
		l.Do("fetchOrders", fetchOrders, UseRun("userID"))
		l.Do("fetchSettings", fetchSettings, UseRun("userID"))
		l.Do("generateReport", generateReport, Use("fetchUser"), Use("fetchOrders"))
		l.Do("processUserData", processUserData, Use("fetchUser"), Use("fetchSettings"))

		_, err := l.Run(context.Background(), map[string]any{"userID": 123})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNestedFieldAccess tests performance of nested field extraction.
func BenchmarkNestedFieldAccess(b *testing.B) {
	for range b.N {
		l := New()
		l.Do("fetchUser", fetchUser, UseRun("userID"))
		l.Do("fetchOrders", fetchOrders, UseRun("userID"))
		l.Do("processAddress", func(ctx context.Context, street, city string) (string, error) {
			return fmt.Sprintf("%s, %s", street, city), nil
		}, Use("fetchUser", "Address", "Street"), Use("fetchUser", "Address", "City"))

		_, err := l.Run(context.Background(), map[string]any{"userID": 123})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManualCoordination compares against manual goroutine coordination.
// compare against BenchmarkSimpleDAG.
//
//revive:disable-next-line:cognitive-complexity,unchecked-type-assertion
func BenchmarkManualCoordination(b *testing.B) {
	for range b.N {
		ctx := context.Background()
		userID := 123

		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make(map[string]any)
		errChan := make(chan error, 2)

		// Start parallel tasks
		wg.Add(2)
		go func() {
			defer wg.Done()
			user, err := fetchUser(ctx, userID)
			if err != nil {
				errChan <- err
				return
			}
			mu.Lock()
			results["user"] = user
			mu.Unlock()
		}()

		go func() {
			defer wg.Done()
			orders, err := fetchOrders(ctx, userID)
			if err != nil {
				errChan <- err
				return
			}
			mu.Lock()
			results["orders"] = orders
			mu.Unlock()
		}()

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			if err != nil {
				b.Fatal(err)
			}
		}

		// Generate report
		mu.Lock()
		//revive:disable-next-line:unchecked-type-assertion
		user := results["user"].(User)
		//revive:disable-next-line:unchecked-type-assertion
		orders := results["orders"].([]Order)
		mu.Unlock()

		_, err := generateReport(ctx, user, orders)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDAGConstruction tests just the DAG building performance
func BenchmarkDAGConstruction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := New()
		l.Do("fetchUser", fetchUser, UseRun("userID"))
		l.Do("fetchOrders", fetchOrders, UseRun("userID"))
		l.Do("fetchSettings", fetchSettings, UseRun("userID"))
		l.Do("generateReport", generateReport, Use("fetchUser"), Use("fetchOrders"))
		l.Do("processUserData", processUserData, Use("fetchUser"), Use("fetchSettings"))
	}
}

// BenchmarkDAGValidation tests validation performance
func BenchmarkDAGValidation(b *testing.B) {
	l := New()
	l.Do("fetchUser", fetchUser, UseRun("userID"))
	l.Do("fetchOrders", fetchOrders, UseRun("userID"))
	l.Do("fetchSettings", fetchSettings, UseRun("userID"))
	l.Do("generateReport", generateReport, Use("fetchUser"), Use("fetchOrders"))
	l.Do("processUserData", processUserData, Use("fetchUser"), Use("fetchSettings"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test just the stage generation (which includes validation)
		stages, err := l.getStages()
		if err != nil {
			b.Fatal(err)
		}
		_ = stages
	}
}

// BenchmarkCPUIntensive tests CPU-bound tasks
func BenchmarkCPUIntensive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := New()
		l.Do("cpu1", cpuIntensiveTask, UseRun("iterations"))
		l.Do("cpu2", cpuIntensiveTask, UseRun("iterations"))
		l.Do("cpu3", cpuIntensiveTask, UseRun("iterations"))
		l.Do("sum", func(ctx context.Context, a, b, c int) (int, error) {
			return a + b + c, nil
		}, Use("cpu1"), Use("cpu2"), Use("cpu3"))

		_, err := l.Run(context.Background(), map[string]any{"iterations": 1000})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryIntensive tests memory allocation patterns
func BenchmarkMemoryIntensive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := New()
		l.Do("mem1", memoryIntensiveTask, UseRun("size"))
		l.Do("mem2", memoryIntensiveTask, UseRun("size"))
		l.Do("concat", func(ctx context.Context, a, b []byte) ([]byte, error) {
			result := make([]byte, len(a)+len(b))
			copy(result, a)
			copy(result[len(a):], b)
			return result, nil
		}, Use("mem1"), Use("mem2"))

		_, err := l.Run(context.Background(), map[string]any{"size": 1024})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScalabilityTest tests with many small tasks
func BenchmarkScalabilityTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := New()

		// Create 10 independent tasks
		for j := 0; j < 10; j++ {
			taskID := fmt.Sprintf("task%d", j)
			l.Do(taskID, func(ctx context.Context, input int) (int, error) {
				return input * 2, nil
			}, UseRun("input"))
		}

		// Create a final task that depends on all others
		inputs := make([]internal.InputSpec, 10)
		for j := 0; j < 10; j++ {
			inputs[j] = Use(fmt.Sprintf("task%d", j))
		}

		l.Do("final", func(ctx context.Context, ip1, ip2, ip3, ip4, ip5, ip6, ip7, ip8, ip9, ip10 int) (int, error) {

			sum := 0
			for _, v := range []int{ip1, ip2, ip3, ip4, ip5, ip6, ip7, ip8, ip9, ip10} {
				sum += v
			}
			return sum, nil
		}, inputs...)

		_, err := l.Run(context.Background(), map[string]any{"input": 42})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark with memory and allocation tracking
func BenchmarkWithMemStats(b *testing.B) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New()
		l.Do("fetchUser", fetchUser, UseRun("userID"))
		l.Do("fetchOrders", fetchOrders, UseRun("userID"))
		l.Do("generateReport", generateReport, Use("fetchUser"), Use("fetchOrders"))

		_, err := l.Run(context.Background(), map[string]any{"userID": 123})
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "mallocs/op")
}

// Test for lock contention with many goroutines
func BenchmarkLockContention(b *testing.B) {
	l := New()
	l.Do("shared", func(ctx context.Context) (int, error) {
		time.Sleep(time.Microsecond) // Small delay to increase contention
		return 42, nil
	})

	// Multiple tasks that depend on the shared task
	for i := 0; i < 10; i++ {
		taskID := fmt.Sprintf("consumer%d", i)
		l.Do(taskID, func(ctx context.Context, input int) (int, error) {
			return input + 1, nil
		}, Use("shared"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := l.Run(context.Background(), map[string]any{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Parallel benchmark to test concurrent usage
func BenchmarkParallelExecution(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l := New()
			l.Do("fetchUser", fetchUser, UseRun("userID"))
			l.Do("fetchOrders", fetchOrders, UseRun("userID"))
			l.Do("generateReport", generateReport, Use("fetchUser"), Use("fetchOrders"))

			_, err := l.Run(context.Background(), map[string]any{"userID": 123})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Benchmark different DAG sizes
func BenchmarkDAGSizes(b *testing.B) {
	sizes := []int{5, 10, 25, 50}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				l := New()

				// Create a chain of tasks
				l.Do("task0", func(ctx context.Context) (int, error) {
					return 1, nil
				})

				for j := 1; j < size; j++ {
					prevTask := fmt.Sprintf("task%d", j-1)
					currentTask := fmt.Sprintf("task%d", j)
					l.Do(currentTask, func(ctx context.Context, input int) (int, error) {
						return input + 1, nil
					}, Use(prevTask))
				}

				_, err := l.Run(context.Background(), map[string]any{})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark functions.
func fetchUser(ctx context.Context, userID int) (User, error) {
	return User{
		ID:    userID,
		Name:  fmt.Sprintf("User%d", userID),
		Email: fmt.Sprintf("user%d@example.com", userID),
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
			State:  "CA",
		},
	}, nil
}

func fetchOrders(ctx context.Context, userID int) ([]Order, error) {
	orders := make([]Order, 3)
	for i := range orders {
		orders[i] = Order{
			ID:     i + 1,
			UserID: userID,
			Amount: float64((i+1)*100 + 50),
			Items: []Item{
				{ID: i*2 + 1, Name: fmt.Sprintf("Item%d", i*2+1), Price: 25.99, Quantity: 1},
				{ID: i*2 + 2, Name: fmt.Sprintf("Item%d", i*2+2), Price: 15.99, Quantity: 2},
			},
		}
	}
	return orders, nil
}

//nolint:gocritic // This is test case.
func generateReport(ctx context.Context, user User, orders []Order) (Report, error) {
	var totalSpent float64
	for _, order := range orders {
		totalSpent += order.Amount
	}

	return Report{
		UserID:     user.ID,
		UserName:   user.Name,
		TotalSpent: totalSpent,
		OrderCount: len(orders),
	}, nil
}

func fetchSettings(ctx context.Context, userID int) (map[string]string, error) {
	return map[string]string{
		"theme":         "dark",
		"language":      "en",
		"timezone":      "UTC",
		"notifications": "enabled",
	}, nil
}

//nolint:gocritic // This is test case.
func processUserData(ctx context.Context, user User, settings map[string]string) (string, error) {
	return fmt.Sprintf("Processed %s with theme %s", user.Name, settings["theme"]), nil
}

// CPU-intensive task for testing
func cpuIntensiveTask(ctx context.Context, iterations int) (int, error) {
	sum := 0
	for i := 0; i < iterations; i++ {
		sum += i * i
	}
	return sum, nil
}

// Memory-intensive task
func memoryIntensiveTask(ctx context.Context, size int) ([]byte, error) {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data, nil
}
