# Lyra - Type-Safe DAG Task Orchestration Library

[![Go Reference](https://pkg.go.dev/badge/github.com/sourabh-kumar2/lyra.svg)](https://pkg.go.dev/github.com/sourabh-kumar2/lyra)
[![Go Report Card](https://goreportcard.com/badge/github.com/sourabh-kumar2/lyra)](https://goreportcard.com/report/github.com/sourabh-kumar2/lyra)

Lyra is a Go library for orchestrating dependent tasks in a Directed Acyclic Graph (DAG) with automatic concurrency and type-safe result passing between tasks.

## Problem Solved

Replace manual `sync.WaitGroup` + channel coordination for complex task dependencies with a clean, type-safe API that:

- ✅ **Eliminates boilerplate**: No manual goroutine/channel management
- ✅ **Ensures type safety**: Runtime validation with excellent error messages
- ✅ **Maximizes concurrency**: Tasks run in parallel when dependencies allow
- ✅ **Prevents deadlocks**: Built-in cycle detection and validation

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/sourabh-kumar2/lyra"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type Report struct {
    UserID   int    `json:"user_id"`
    UserName string `json:"user_name"`
    Summary  string `json:"summary"`
}

func fetchUser(ctx context.Context, userID int) (User, error) {
    // Simulate API call
    return User{ID: userID, Name: "John Doe"}, nil
}

func generateReport(ctx context.Context, user User) (Report, error) {
    return Report{
        UserID:   user.ID,
        UserName: user.Name,
        Summary:  fmt.Sprintf("Report for user %s", user.Name),
    }, nil
}

func main() {
    l := lyra.New()
    
    // Define tasks with explicit input mapping
    l.Do("fetchUser", fetchUser, lyra.UseRun("userID"))
    l.Do("generateReport", generateReport, lyra.Use("fetchUser"))
    
    // Execute with runtime inputs
    results, err := l.Run(context.Background(), map[string]any{
        "userID": 123,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Access results (type assertions required)
    user, _ := results.Get("fetchUser")
    report, _ := results.Get("generateReport")
    
    fmt.Printf("User: %+v\n", user.(User))
    fmt.Printf("Report: %+v\n", report.(Report))
}
```

## Core Features

### 🔗 Dependency Management
Tasks automatically execute in the correct order based on their dependencies:

```go
l := lyra.New()
l.Do("task1", func(ctx context.Context) (string, error) { return "result1", nil })
l.Do("task2", func(ctx context.Context, input string) (string, error) { 
    return input + "_processed", nil 
}, lyra.Use("task1"))
```

### 🚀 Automatic Concurrency
Independent tasks run in parallel automatically:

```go
// These run concurrently since they have no dependencies between them
l.Do("fetchUser", fetchUserFunc, lyra.UseRun("userID"))
l.Do("fetchSettings", fetchSettingsFunc, lyra.UseRun("userID"))

// This waits for both to complete
l.Do("merge", mergeFunc, lyra.Use("fetchUser"), lyra.Use("fetchSettings"))
```

### 🎯 Nested Field Access
Extract specific fields from task results:

```go
type UserResponse struct {
    User User `json:"user"`
    Meta Meta `json:"meta"`
}

// Access nested fields directly
l.Do("processUser", processFunc, lyra.Use("fetchUser", "User", "Name"))
```

### ⚡ Runtime Input Injection
Provide initial values at execution time:

```go
results, err := l.Run(ctx, map[string]any{
    "apiKey":    "secret-key",
    "userID":    123,
    "batchSize": 50,
})
```

## Function Requirements

All task functions must follow these signature patterns:

```go
// Pattern 1: Context only, no output
func(context.Context) error

// Pattern 2: Context only, with output
func(context.Context) (ResultType, error)  

// Pattern 3: Context + inputs, with output
func(context.Context, input1 Type1, input2 Type2) (ResultType, error)
```

**Requirements:**
- ✅ First parameter must be `context.Context`
- ✅ Last return value must be `error`
- ✅ Can return `(result, error)` or just `(error)`
- ❌ Variadic functions not supported

## Input Specification

### Use() - Task Dependencies
Reference results from other tasks:

```go
// Use entire result
lyra.Use("taskID")

// Use specific field  
lyra.Use("taskID", "FieldName")

// Use nested field
lyra.Use("taskID", "User", "Address", "Street")
```

### UseRun() - Runtime Inputs
Reference values provided at runtime:

```go
// Use runtime value
lyra.UseRun("configKey")

// Use nested field from runtime value
lyra.UseRun("config", "Database", "Host")
```

## Error Handling

Lyra provides excellent error messages for common issues:

```go
// Missing dependency
l.Do("task2", func(ctx context.Context, input string) error { return nil }, 
     lyra.Use("nonexistent"))
// Error: dependency not found: node "task2" depends on non-existent node "nonexistent"

// Type mismatch  
l.Do("task2", func(ctx context.Context, input int) error { return nil },
     lyra.Use("task1")) // task1 returns string
// Error: parameter 2 -> expected type int, got string

// Cyclic dependency
l.Do("task1", func(ctx context.Context, input string) (string, error) { return "", nil }, 
     lyra.Use("task2"))
l.Do("task2", func(ctx context.Context, input string) (string, error) { return "", nil }, 
     lyra.Use("task1"))  
// Error: cyclic dependency detected
```

## Advanced Examples

### Data Pipeline
```go
func buildDataPipeline() *lyra.Lyra {
    l := lyra.New()
    
    // Extract data from multiple sources
    l.Do("extractUsers", extractUsers, lyra.UseRun("dbConfig"))
    l.Do("extractOrders", extractOrders, lyra.UseRun("dbConfig"))
    l.Do("extractProducts", extractProducts, lyra.UseRun("apiConfig"))
    
    // Transform data (runs in parallel)
    l.Do("cleanUsers", cleanUsers, lyra.Use("extractUsers"))
    l.Do("enrichOrders", enrichOrders, 
         lyra.Use("extractOrders"), 
         lyra.Use("extractProducts"))
    
    // Load final result
    l.Do("generateReport", generateReport, 
         lyra.Use("cleanUsers"), 
         lyra.Use("enrichOrders"))
         
    return l
}
```

### Microservice Initialization
```go
func initializeServices() *lyra.Lyra {
    l := lyra.New()
    
    // Independent initialization
    l.Do("initDB", initDatabase, lyra.UseRun("dbConfig"))
    l.Do("initCache", initCache, lyra.UseRun("cacheConfig"))
    l.Do("loadConfig", loadConfiguration, lyra.UseRun("configPath"))
    
    // Dependent initialization
    l.Do("setupMetrics", setupMetrics, lyra.Use("loadConfig", "MetricsConfig"))
    l.Do("startServer", startHTTPServer, 
         lyra.Use("initDB"), 
         lyra.Use("initCache"),
         lyra.Use("loadConfig", "ServerConfig"))
         
    return l
}
```

## Performance

- **Concurrency**: Tasks execute in parallel when dependencies allow
- **Efficiency**: No unnecessary goroutine creation for single tasks
- **Memory**: Results stored only until all dependent tasks complete
- **Validation**: All validation occurs upfront, not during execution

## Comparison

### Without Lyra (Manual Coordination)
```go
func processManually(ctx context.Context, userID int) error {
    var wg sync.WaitGroup
    var mu sync.Mutex
    results := make(map[string]interface{})
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
            return err
        }
    }
    
    // Generate report with results
    user := results["user"].(User)
    orders := results["orders"].([]Order)
    _, err := generateReport(ctx, user, orders)
    return err
}
```

### With Lyra (Clean & Safe)
```go
func processWithLyra(ctx context.Context, userID int) error {
    l := lyra.New()
    l.Do("fetchUser", fetchUser, lyra.UseRun("userID"))
    l.Do("fetchOrders", fetchOrders, lyra.UseRun("userID"))
    l.Do("generateReport", generateReport, 
         lyra.Use("fetchUser"), 
         lyra.Use("fetchOrders"))
    
    _, err := l.Run(ctx, map[string]any{"userID": userID})
    return err
}
```

## Installation

```bash
go get github.com/sourabh-kumar2/lyra
```

## Requirements

- Go 1.23+
- No external dependencies

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the AGPL-3.0 License - see the [LICENSE](LICENSE) file for details.
