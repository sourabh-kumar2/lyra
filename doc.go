// Package lyra provides a type-safe DAG task orchestration library for coordinating
// dependent tasks with automatic concurrency and result passing.
//
// Lyra replaces manual sync.WaitGroup and channel coordination with a clean API
// that ensures type safety, maximizes concurrency, and prevents common pitfalls
// like deadlocks and race conditions.
//
// Basic usage:
//
//	l := lyra.New()
//	l.Do("task1", func(ctx context.Context) (string, error) { return "result", nil })
//	l.Do("task2", func(ctx context.Context, input string) error { return nil }, lyra.Use("task1"))
//
//	results, err := l.Run(context.Background(), map[string]any{})
package lyra
