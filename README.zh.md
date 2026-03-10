[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/mutexmap/release.yml?branch=main&label=BUILD)](https://github.com/yylego/mutexmap/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/mutexmap)](https://pkg.go.dev/github.com/yylego/mutexmap)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/mutexmap/main.svg)](https://coveralls.io/github/yylego/mutexmap?branch=main)
![Supported Go Versions](https://img.shields.io/badge/Go-1.22--1.25-lightgrey.svg)
[![GitHub Release](https://img.shields.io/github/release/yylego/mutexmap.svg)](https://github.com/yylego/mutexmap/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/mutexmap)](https://goreportcard.com/report/github.com/yylego/mutexmap)

# MutexMap - 基于 Mutex 实现的线程安全 Map

使用 `sync.RWMutex` 实现的线程安全 map。专门用于处理并发读写场景，在多线程应用中提供高效可靠的操作。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->
## 英文文档

[ENGLISH README](README.md)
<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 项目概览

Go 标准库的 `map` 在并发访问时不安全。本包使用 `sync.RWMutex` 封装 `map[K]V` 以提供线程安全。通过 `RWMutex`，多个读取操作可以同时访问 map，而写入操作会被同步以防止竞态条件。

### 主要特性

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

⬆️ **源码:** [源码](internal/demos/demo2x/main.go)

---

## 功能概览

### 方法参考

| 方法                                     | 描述                                         |
|----------------------------------------|--------------------------------------------|
| `NewMap[K comparable, V any](cap int)` | 创建新的 `Map`，带起始容量提示。                        |
| `Get(k K) (V, bool)`                   | 获取与 ID 关联的值。ID 不存在时返回 `false`。             |
| `Set(k K, v V)`                        | 将值与给定 ID 关联。如果 ID 存在则覆盖。                  |
| `Delete(k K)`                          | 从 map 中删除关联。                              |
| `Len() int`                            | 返回 map 中关联的数量。                            |
| `Range(func(k K, v V) bool)`           | 遍历每个关联。当函数返回 `false` 时停止。                 |
| `Getset(k K, func() V) (V, enum)`      | 获取值，当值不存在时设置它，防止重复计算。                    |

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
db, status := mp.Getset("primary", func() *Database {
    return ConnectDatabase("primary")
})
```

### 安全初始化模式

**按需配置加载：**
```go
configMap := mutexmap.NewMap[string, Config](10)

config, status := configMap.Getset("app", func() Config {
    return LoadConfigFromFile("config.json")
})
// 第一次调用：status == mutexmap.CacheSet
// 后续调用：status == mutexmap.CacheGet
```

**数据库连接池：**
```go
connMap := mutexmap.NewMap[string, *sql.DB](5)

conn, _ := connMap.Getset("mysql", func() *sql.DB {
    db, _ := sql.Open("mysql", "user:pass@/dbname")
    return db
})
```

### 检查-执行操作

**检查后更新：**
```go
mp := mutexmap.NewMap[string, int](10)
mp.Set("score", 100)

if value, found := mp.Get("score"); found {
    mp.Set("score", value+10) // 增加现有值
}
```

**存在时删除：**
```go
if _, found := mp.Get("tempData"); found {
    mp.Delete("tempData")
}
```

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

response, status := cache.Getset(requestURL, func() APIResponse {
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

session, status := sessions.Getset(sessionID, func() Session {
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
<!-- VERSION 2025-09-26 07:39:27.188023 +0000 UTC -->

## 📄 许可证类型

MIT 许可证。详见 [LICENSE](LICENSE)。

---

## 🤝 项目贡献

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **发现问题？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **功能建议？** 创建 issue 讨论您的想法
- 📖 **文档疑惑？** 报告问题，帮助我们改进文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，帮助我们优化性能
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
7. **文档**：为面向用户的更改更新文档，并使用有意义的提交消息
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来为此项目做出贡献。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->
