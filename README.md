[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/mutexmap/release.yml?branch=main&label=BUILD)](https://github.com/yylego/mutexmap/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/mutexmap)](https://pkg.go.dev/github.com/yylego/mutexmap)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/mutexmap/main.svg)](https://coveralls.io/github/yylego/mutexmap?branch=main)
![Supported Go Versions](https://img.shields.io/badge/Go-1.22--1.25-lightgrey.svg)
[![GitHub Release](https://img.shields.io/github/release/yylego/mutexmap.svg)](https://github.com/yylego/mutexmap/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/mutexmap)](https://goreportcard.com/report/github.com/yylego/mutexmap)

# MutexMap - A Thread-Safe Map for Go

A thread-safe map implementation using `sync.RWMutex` to synchronize access. This package is optimized to handle concurrent reads and writes, providing efficient and reliable operations in multi-thread applications.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->
## CHINESE README

[中文说明](README.zh.md)
<!-- TEMPLATE (EN) END: LANGUAGE NAVIGATION -->

## Overview

Go's standard `map` is not safe when accessed concurrently. This package wraps a `map[K]V` with `sync.RWMutex` to provide thread safety. With `RWMutex`, multiple readers can access the map at once, while writes are synchronized to prevent race conditions.

Main highlights:

- **Thread-safety**: Prevents data races in concurrent environments.
- **Efficient Reads**: Read operations (`Get`, `Range`) allow simultaneous access without blocking.
- **Synchronous Write**: Write operations (`Set`, `Delete`, `Getset`) are synchronous.

This package is suitable when you need frequent reads and infrequent writes with a shared map.

## Installation

```bash  
go get github.com/yylego/mutexmap  
```  

## Example Usage

### Basic Operations

```go
package main

import (
	"fmt"
	"github.com/yylego/mutexmap"
)

func main() {
	mp := mutexmap.NewMap[string, int](10)

	mp.Set("key1", 100)
	mp.Set("key2", 200)

	if value, found := mp.Get("key1"); found {
		fmt.Println("Key1 Value:", value)
	}

	mp.Range(func(key string, value int) bool {
		fmt.Println(key, value)
		return true
	})
}
```

⬆️ **Source:** [Source](internal/demos/demo1x/main.go)

### Using `Getset` for Cached Initialization

```go
package main

import (
	"fmt"

	"github.com/yylego/mutexmap"
)

func main() {
	mp := mutexmap.NewMap[string, string](10)

	value, status := mp.Getset("exampleKey", func() string {
		return "This is a computed value"
	})
	fmt.Println("Status:", status, "Value:", value)

	value, status = mp.Getset("exampleKey", func() string {
		return "New value"
	})
	fmt.Println("Status:", status, "Value:", value)
}
```

⬆️ **Source:** [Source](internal/demos/demo2x/main.go)

## Features

- **Concurrent Access**: Allows multiple goroutines to read and write to the map with thread protection.
- **Optimized Reads**: Supports simultaneous reads to improve performance.
- **Custom Initialization**: Use `Getset` to set a new value when the given ID does not exist.

---

## Method Reference

| Method                                 | Description                                                                              |
|----------------------------------------|------------------------------------------------------------------------------------------|
| `NewMap[K comparable, V any](cap int)` | Creates a new `Map` with starting capacity hint.                                         |
| `Get(k K) (V, bool)`                   | Retrieves the value associated with the ID. Returns `false` when the ID does not exist.  |
| `Set(k K, v V)`                        | Associates a value with the given ID. Overwrites if the ID exists.                       |
| `Delete(k K)`                          | Removes the association from the map.                                                    |
| `Len() int`                            | Returns the count of associations in the map.                                            |
| `Range(func(k K, v V) bool)`           | Iterates through each association. Stops when the function returns `false`.              |
| `Getset(k K, func() V) (V, enum)`      | Gets a value and sets it when it doesn't exist, preventing duplicate computation.        |  

---

## Examples

This section demonstrates various usage patterns and techniques with MutexMap.

### Concurrent Access Patterns

**Safe concurrent reads and writes:**
```go
mp := mutexmap.NewMap[string, int](10)

// Multiple goroutines can access the map with thread protection
go func() {
    mp.Set("counter", 100)
}()

go func() {
    if value, ok := mp.Get("counter"); ok {
        fmt.Println("Value:", value)
    }
}()
```

**Concurrent initialization with Getset:**
```go
mp := mutexmap.NewMap[string, *Database](10)

// Multiple goroutines requesting the same resource
// Just one creates it, the rest wait and receive the same instance
db, status := mp.Getset("primary", func() *Database {
    return ConnectDatabase("primary")
})
```

### Safe Initialization Patterns

**On-demand configuration loading:**
```go
configMap := mutexmap.NewMap[string, Config](10)

config, status := configMap.Getset("app", func() Config {
    return LoadConfigFromFile("config.json")
})
// First call: status == mutexmap.CacheSet
// Subsequent calls: status == mutexmap.CacheGet
```

**Database connection pooling:**
```go
connMap := mutexmap.NewMap[string, *sql.DB](5)

conn, _ := connMap.Getset("mysql", func() *sql.DB {
    db, _ := sql.Open("mysql", "user:pass@/dbname")
    return db
})
```

### Check-And-Act Operations

**Check before update:**
```go
mp := mutexmap.NewMap[string, int](10)
mp.Set("score", 100)

if value, found := mp.Get("score"); found {
    mp.Set("score", value+10) // Increment existing value
}
```

**Delete if exists:**
```go
if _, found := mp.Get("tempData"); found {
    mp.Delete("tempData")
}
```

### Iteration Patterns

**Process each item:**
```go
mp.Range(func(key string, value int) bool {
    fmt.Printf("%s: %d\n", key, value)
    return true // Continue iteration
})
```

**Find specific entry:**
```go
mp.Range(func(key string, value int) bool {
    if value > 100 {
        fmt.Println("Found:", key)
        return false // Stop iteration
    }
    return true
})
```

**Count matching entries:**
```go
count := 0
mp.Range(func(key string, value int) bool {
    if value > threshold {
        count++
    }
    return true
})
```

### Type-Safe Usage

**String to struct mapping:**
```go
type Account struct {
    ID   int
    Name string
}

accountMap := mutexmap.NewMap[string, Account](10)
accountMap.Set("user1", Account{ID: 1, Name: "Alice"})
```

**Int to function mapping:**
```go
handlerMap := mutexmap.NewMap[int, func() error](10)
handlerMap.Set(404, func() error {
    return fmt.Errorf("not found")
})
```

### Cache Use Cases

**API response caching:**
```go
cache := mutexmap.NewMap[string, APIResponse](100)

response, status := cache.Getset(requestURL, func() APIResponse {
    return FetchFromAPI(requestURL)
})
```

**Computed value caching:**
```go
resultCache := mutexmap.NewMap[string, int](50)

result, _ := resultCache.Getset("expensive-computation", func() int {
    return PerformExpensiveCalculation()
})
```

**Session management:**
```go
sessions := mutexmap.NewMap[string, Session](100)

session, status := sessions.Getset(sessionID, func() Session {
    return CreateNewSession(sessionID)
})
```

### Batch Operations

**Bulk insert:**
```go
data := map[string]int{"a": 1, "b": 2, "c": 3}
mp := mutexmap.NewMap[string, int](len(data))

for k, v := range data {
    mp.Set(k, v)
}
```

**Bulk check:**
```go
keys := []string{"key1", "key2", "key3"}
found := make(map[string]bool)

for _, key := range keys {
    if _, ok := mp.Get(key); ok {
        found[key] = true
    }
}
```

---

## Benefits

1. **Thread Protection**: Required when sharing maps across multiple threads.
2. **Efficient Reads**: Read lock (`RLock`) allows simultaneous read access without contention.
3. **Write Protection**: Write lock (`Lock`) maintains data coherence during changes.
4. **Smart Initialization**: The `Getset` method avoids redundant computation.

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-09-26 07:39:27.188023 +0000 UTC -->

## 📄 License

MIT License. See [LICENSE](LICENSE).

---

## 🤝 Contributing

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Found a mistake?** Open an issue on GitHub with reproduction steps
- 💡 **Have a feature idea?** Create an issue to discuss the suggestion
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes and use significant commit messages
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) END: STANDARD PROJECT FOOTER -->

## GitHub Stars

[![starring](https://starchart.cc/yylego/mutexmap.svg?variant=adaptive)](https://starchart.cc/yylego/mutexmap)
