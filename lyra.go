package lyra

// New creates a new Lyra instance for building and executing DAGs.
//
//	l := lyra.New()
//	l.Do("task1", taskFunc1).Do("task2", taskFunc2).After("task1")
func New() *Lyra {
	return &Lyra{}
}

// Lyra coordinates dependent tasks that can run concurrently when possible,
// with compile-time type safety for result passing between tasks.
// It replaces manual sync.WaitGroup and channel
// coordination with a clean, fluent API.
type Lyra struct{}
