<!-- TEMPLATE (ZH) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/mutexmap/release.yml?branch=main&label=BUILD)](https://github.com/yylego/mutexmap/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/mutexmap)](https://pkg.go.dev/github.com/yylego/mutexmap)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/mutexmap/main.svg)](https://coveralls.io/github/yylego/mutexmap?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/mutexmap.svg)](https://github.com/yylego/mutexmap/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/mutexmap)](https://goreportcard.com/report/github.com/yylego/mutexmap)

<!-- TEMPLATE (ZH) CLOSE: BADGES -->

# MutexMap - 基于 Mutex 实现的线程安全 Map

使用 `sync.RWMutex` 实现的线程安全 map。专门用于处理并发读写场景，在多线程应用中提供高效可靠的操作。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) CLOSE: LANGUAGE NAVIGATION -->

## 项目概览

Go 标准库的 `map` 在并发访问时不安全。本包使用 `sync.RWMutex` 封装 `map[K]V` 以提供线程安全。通过 `RWMutex`，多个读取操作可以同时访问 map，而写入操作会被同步以防止竞态条件。

主要亮点：

- **线程安全**：避免并发环境中的数据竞争。
- **高效读取**：读取操作（`Get`、`Range`）支持同时访问而不阻塞。
- **同步写入**：写入操作（`Set`、`Delete`、`Getset`）是同步的。

本包适用于需要频繁读取和偶尔写入共享 map 的场景。

---

## 安装

使用 `go get` 安装本库：

```bash
go get github.com/yylego/mutexmap
```

---

## 示例代码

### 基础操作示例

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

⬆️ **源码:** [源码](internal/demos/demo1x/main.go)

### 使用 `Getset` 方法实现值的缓存初始化

```go
package main

import (
	"fmt"

	"github.com/yylego/mutexmap"
)

func main() {
	mp := mutexmap.NewMap[string, string](10)

	// 第一次调用：值不存在，计算后存进去，loaded 为 false
	value, loaded := mp.Getset("exampleID", func() string {
		return "This is a computed value"
	})
	fmt.Println("Loaded:", loaded, "Value:", value)

	// 第二次调用：值已经在 map 里，函数不会执行，loaded 为 true
	value, loaded = mp.Getset("exampleID", func() string {
		return "New value"
	})
	fmt.Println("Loaded:", loaded, "Value:", value)
}
```

⬆️ **源码:** [源码](internal/demos/demo2x/main.go)

---

## 功能特性

- **并发访问**：多个协程可以在线程保护下读写 map。
- **读取优化**：支持同时读取以提升性能。
- **自定义初始化**：用 `Getset` 在给定 ID 不存在时设置新值。

---

## 方法参考

| 方法                                    | 描述                                                                                                 |
| --------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `NewMap[K comparable, V any](size int)` | 创建新的 `Map`，带起始容量提示。                                                                     |
| `Get(k K) (V, bool)`                    | 获取与 ID 关联的值。ID 不存在时返回 `false`。                                                        |
| `Set(k K, v V)`                         | 将值与给定 ID 关联。如果 ID 存在则覆盖。                                                             |
| `Delete(k K)`                           | 从 map 中删除关联。                                                                                  |
| `Len() int`                             | 返回 map 中关联的数量。                                                                              |
| `Range(func(k K, v V) bool)`            | 遍历每个关联。当函数返回 `false` 时停止。                                                            |
| `Getset(k K, func() V) (V, bool)`       | 获取值，不存在时计算并设置；值本来就在 map 里时 `loaded` 为 `true`，跟 `sync.Map.LoadOrStore` 一样。 |

`Map` 的零值可以直接使用，跟 `sync.Map` 一样；`New` 和 `NewMap` 只是设置初始容量。

`Range` 在回调运行期间一直拿着读锁，所以在回调里对同一个 map 调用 `Set`、`Delete` 或 `Getset` 会永远卡住。先把要处理的键收集起来，等 `Range` 返回后再改 map。

`Getset` 在整个 map 的写锁里计算新值，计算慢的话，这期间其他键也都被挡住。计算慢或者可能失败时，用 `cachemap`（见下一节）。

## 带错误的按需缓存（`cachemap`）

`cachemap.Map` 跟 `Getset` 一样每个键只计算一次，区别有三点：计算慢的时候不挡住其他键；计算可以返回错误；错误也会被缓存。

```go
connCache := cachemap.New[string, *sql.DB]()

db, err := connCache.Getset("mysql", func() (*sql.DB, error) {
    return sql.Open("mysql", "user:pass@/dbname")
})
if err != nil {
    return err // 错误也会被缓存，想重试就调用 connCache.Delete("mysql")
}
```

- 多个协程同时取同一个键时，只在这个键上等待，计算只执行一次。
- 计算里发生的 panic 会作为这个键的错误返回。
- `Get(k)` 返回 `res, ok, err`：键不存在、或还在计算中时 `ok` 为 `false`，这时 `err` 为 `nil`。「没找到」不算错误，`err` 只是计算返回的错误。

## 与标准库的对照

Go 官方提供了相近的工具，我们推荐先看看它们：官方维护、文档最全，学会了在每个 Go 项目里都用得上。本包照常可用、继续保留。

- **`sync.Map`**（标准库 `sync` 包）：只需要一个并发安全的 map、不介意写类型断言时，用它就够了。`mutexmap.Map` 多出来的是泛型、`Len`，以及 `Getset` 只在键不存在时才执行计算函数，而 `LoadOrStore` 收的是事先算好的值。想保留 `sync.Map` 的用法、只是不想写类型断言，见 [syncmap](https://github.com/yylego/syncmap)，它是 `sync.Map` 的泛型封装。
- **`sync.OnceValues` 配 `sync.Map`**（标准库 `sync` 包，`OnceValues` 要 Go 1.21 起）：两者拼起来就是「每个键只算一次，错误也一起缓存」：

  ```go
  var cache sync.Map // 键 -> func() (*sql.DB, error)

  load, _ := cache.LoadOrStore("mysql", sync.OnceValues(func() (*sql.DB, error) {
      return sql.Open("mysql", "user:pass@/dbname")
  }))
  db, err := load.(func() (*sql.DB, error))()
  ```

  `cachemap.Map` 多出来三样：不用写类型断言；计算里的 panic 会转成错误返回（`OnceValues` 每次调用都会把 panic 再抛一遍）；可以 `Delete` 之后重新计算。syncmap 里的 [oncemap](https://github.com/yylego/syncmap) 就是把这两块标准库积木包好了的版本，用法跟 `sync.Map` 一个风格。

- **`golang.org/x/sync/singleflight`**（Go 官方的扩展库）：只想让「同时进来的请求只算一次」、算完不需要缓存时，用它就够了。

---

## 使用示例

本节展示 MutexMap 的各种使用模式和技巧。

### 并发访问模式

**安全的并发读写：**

```go
mp := mutexmap.NewMap[string, int](10)

// 多个协程可以安全地访问 map
go func() {
    mp.Set("counter", 100)
}()

go func() {
    if value, ok := mp.Get("counter"); ok {
        fmt.Println("Value:", value)
    }
}()
```

**使用 Getset 并发初始化：**

```go
mp := mutexmap.NewMap[string, *Database](10)

// 多个协程请求同一个资源
// 只有一个会创建它，其他的会等待并获取同一个实例
db, loaded := mp.Getset("main", func() *Database {
    return ConnectDatabase("main")
})
```

### 安全初始化模式

**按需配置加载：**

```go
configMap := mutexmap.NewMap[string, Config](10)

config, loaded := configMap.Getset("app", func() Config {
    return LoadConfigFromFile("config.json")
})
// 第一次调用：loaded == false，函数执行了
// 后续调用：loaded == true，返回存下的值
```

可能打开失败的连接应该放进 `cachemap`，它会把错误返回出来，见上面一节。

### 先检查后执行不是原子操作

每个方法各自加锁，所以先 `Get` 再 `Set` 是两步，中间可能被别的协程改掉：

```go
// 并发下不安全：两个协程可能都读到 100、都写入 110，丢掉一次加法
if value, found := mp.Get("score"); found {
    mp.Set("score", value+10)
}
```

「不存在才设置」用 `Getset`。「读出来改了再写回去」要自己加一把锁把两步包起来。`Delete` 删不存在的键也是安全的，不用先 `Get` 一次。

### 迭代模式

**处理每个条目：**

```go
mp.Range(func(key string, value int) bool {
    fmt.Printf("%s: %d\n", key, value)
    return true // 继续迭代
})
```

**查找特定条目：**

```go
mp.Range(func(key string, value int) bool {
    if value > 100 {
        fmt.Println("Found:", key)
        return false // 停止迭代
    }
    return true
})
```

**统计符合条件的条目：**

```go
count := 0
mp.Range(func(key string, value int) bool {
    if value > threshold {
        count++
    }
    return true
})
```

### 类型安全使用

**字符串到结构体映射：**

```go
type Account struct {
    ID   int
    Name string
}

accountMap := mutexmap.NewMap[string, Account](10)
accountMap.Set("user1", Account{ID: 1, Name: "Alice"})
```

**整数到函数映射：**

```go
handlerMap := mutexmap.NewMap[int, func() error](10)
handlerMap.Set(404, func() error {
    return fmt.Errorf("not found")
})
```

### 缓存使用场景

**API 响应缓存：**

```go
cache := mutexmap.NewMap[string, APIResponse](100)

response, loaded := cache.Getset(requestURL, func() APIResponse {
    return FetchFromAPI(requestURL)
})
```

**计算结果缓存：**

```go
resultCache := mutexmap.NewMap[string, int](50)

result, _ := resultCache.Getset("expensive-computation", func() int {
    return PerformExpensiveCalculation()
})
```

**会话管理：**

```go
sessions := mutexmap.NewMap[string, Session](100)

session, loaded := sessions.Getset(sessionID, func() Session {
    return CreateNewSession(sessionID)
})
```

### 批量操作

**批量插入：**

```go
data := map[string]int{"a": 1, "b": 2, "c": 3}
mp := mutexmap.NewMap[string, int](len(data))

for k, v := range data {
    mp.Set(k, v)
}
```

**批量检查：**

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

## 优势

1. **线程保护**：在多线程间共享 map 时必需。
2. **高效读取**：读锁（`RLock`）允许同时读取访问而无竞争。
3. **写入保护**：写锁（`Lock`）在变更期间保持数据一致性。
4. **智能初始化**：`Getset` 方法避免冗余计算。

---

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) CLOSE: STANDARD PROJECT FOOTER -->
