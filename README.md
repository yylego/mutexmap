<!-- TEMPLATE (EN) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/mutexmap/release.yml?branch=main&label=BUILD)](https://github.com/yylego/mutexmap/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/mutexmap)](https://pkg.go.dev/github.com/yylego/mutexmap)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/mutexmap/main.svg)](https://coveralls.io/github/yylego/mutexmap?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/mutexmap.svg)](https://github.com/yylego/mutexmap/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/mutexmap)](https://goreportcard.com/report/github.com/yylego/mutexmap)

<!-- TEMPLATE (EN) CLOSE: BADGES -->

# MutexMap - A Thread-Safe Map for Go

A thread-safe map implementation using `sync.RWMutex` to synchronize access. This package is optimized to handle concurrent reads and writes, providing efficient and reliable operations in multi-thread applications.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) CLOSE: LANGUAGE NAVIGATION -->

## Overview

Go's standard `map` is not safe when accessed concurrently. This package wraps a `map[K]V` with `sync.RWMutex` to provide thread safety. With `RWMutex`, multiple readers can access the map at once, while writes are synchronized to prevent race conditions.

Main highlights:

- **Thread-safety**: Prevents data races in concurrent environments.
- **Efficient Reads**: Read operations (`Get`, `Range`) allow simultaneous access without blocking.
- **Synchronous Write**: Write operations (`Set`, `Delete`, `Getset`) are synchronous.

This package is suitable when you need frequent reads and infrequent writes with a shared map.

---

## Installation

Get the package with `go get`:

```bash
go get github.com/yylego/mutexmap
```

---

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

	// First time: the value is absent, so it is computed and stored, loaded is false
	value, loaded := mp.Getset("exampleID", func() string {
		return "This is a computed value"
	})
	fmt.Println("Loaded:", loaded, "Value:", value)

	// Second time: the value is in the map, the function does not run, loaded is true
	value, loaded = mp.Getset("exampleID", func() string {
		return "New value"
	})
	fmt.Println("Loaded:", loaded, "Value:", value)
}
```

⬆️ **Source:** [Source](internal/demos/demo2x/main.go)

---

## Features

- **Concurrent Access**: Allows multiple goroutines to read and write to the map with thread protection.
- **Optimized Reads**: Supports simultaneous reads to improve performance.
- **Custom Initialization**: Use `Getset` to set a new value when the given ID does not exist.

---

## Method Reference

| Method                                  | Description                                                                                                                |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `NewMap[K comparable, V any](size int)` | Creates a new `Map` with a starting size hint.                                                                             |
| `Get(k K) (V, bool)`                    | Retrieves the value associated with the ID. Returns `false` when the ID does not exist.                                    |
| `Set(k K, v V)`                         | Associates a value with the given ID. Overwrites if the ID exists.                                                         |
| `Delete(k K)`                           | Removes the association from the map.                                                                                      |
| `Len() int`                             | Returns the count of associations in the map.                                                                              |
| `Range(func(k K, v V) bool)`            | Iterates through each association. Stops when the function returns `false`.                                                |
| `Getset(k K, func() V) (V, bool)`       | Gets a value and sets it when absent; `loaded` is `true` when the value was in the map, as `sync.Map.LoadOrStore` reports. |

The zero value of `Map` works with no `New`, as with `sync.Map`; `New` and `NewMap` just set the starting size.

`Range` holds the read lock while the function runs, so `Set`, `Delete` and `Getset` on the same map inside it wait without end. Collect the IDs first, and change the map once `Range` returns.

`Getset` computes the new value inside the map-wide write lock, so a slow computation blocks each ID at the same time. Use `cachemap` (see the next section) when the computation is slow, and when it can return an error.

## On-Demand Cache with Errors (`cachemap`)

`cachemap.Map` computes each ID once, as `Getset` does, with three changes: the map stays open to the rest of the IDs during a slow computation, the computation can return an error, and that error is cached too.

```go
connCache := cachemap.New[string, *sql.DB]()

db, err := connCache.Getset("mysql", func() (*sql.DB, error) {
    return sql.Open("mysql", "user:pass@/dbname")
})
if err != nil {
    return err // The error is cached too, and connCache.Delete("mysql") lets the next Getset run the computation again
}
```

- Goroutines asking the same ID at once wait on that ID alone, and the computation runs once.
- A panic inside the computation comes back as the error of that ID.
- `Get(k)` returns `res, ok, err`: `ok` is `false` when the ID is absent and while its computation goes on, and `err` is `nil` then. A miss is not an error, so `err` is just the error the computation gave.

## Standard Package Alternatives

The Go team ships close tools, and we recommend a look at them first: these come with the best documentation, and what you learn there works in each Go project. This package keeps working and stays.

- **`sync.Map`** (standard `sync` package): enough when you need a concurrent map and type assertions are fine with you. `mutexmap.Map` adds generics, `Len`, and a `Getset` that runs the function just when the ID is absent, while `LoadOrStore` takes a value computed in advance. To keep the `sync.Map` API and drop the type assertions, see [syncmap](https://github.com/yylego/syncmap), its generic wrapping.
- **`sync.OnceValues` with `sync.Map`** (standard `sync` package, `OnceValues` since Go 1.21): the two give "compute each ID once, cache the error too":

  ```go
  var cache sync.Map // ID -> func() (*sql.DB, error)

  load, _ := cache.LoadOrStore("mysql", sync.OnceValues(func() (*sql.DB, error) {
      return sql.Open("mysql", "user:pass@/dbname")
  }))
  db, err := load.(func() (*sql.DB, error))()
  ```

  `cachemap.Map` adds three things: no type assertion, a panic coming back as an error (`OnceValues` panics again each time), and `Delete` to run the computation again. [oncemap](https://github.com/yylego/syncmap) in syncmap is these two standard parts wrapped up, in the `sync.Map` API style.

- **`golang.org/x/sync/singleflight`** (the Go team's extension module): enough when goroutines asking the same ID at once need one computation, and the outcome needs no cache.

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
db, loaded := mp.Getset("main", func() *Database {
    return ConnectDatabase("main")
})
```

### Safe Initialization Patterns

**On-demand configuration loading:**

```go
configMap := mutexmap.NewMap[string, Config](10)

config, loaded := configMap.Getset("app", func() Config {
    return LoadConfigFromFile("config.json")
})
// First time: loaded == false, the function ran
// Next times: loaded == true, the stored value comes back
```

Opening a connection can return an error, so it belongs in `cachemap`, which gives the error back, see the section above.

### Check-Then-Act Is Not Atomic

Each method locks on its own, so `Get` then `Set` is two steps, and a different goroutine can change the value in between:

```go
// Unsafe with goroutines at once: two goroutines can both read 100 and both write 110, one increment is lost
if value, found := mp.Get("score"); found {
    mp.Set("score", value+10)
}
```

Use `Getset` to "set when absent". To read, change and write back, guard the two steps with a lock of the calling code. `Delete` is safe on an absent ID, so it needs no `Get` before it.

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

response, loaded := cache.Getset(requestURL, func() APIResponse {
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

session, loaded := sessions.Getset(sessionID, func() Session {
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

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
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
7. **Documentation**: Update documentation to support client-facing changes
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

<!-- TEMPLATE (EN) CLOSE: STANDARD PROJECT FOOTER -->
