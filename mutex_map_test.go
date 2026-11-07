// Package mutexmap_test: Comprehensive test suite for thread-safe map package
// Tests concurrent access patterns, cache behavior, and synchronization correctness
// Validates both basic operations and advanced double-checked locking implementation
//
// mutexmap_test: mutexmap 线程安全 map 包的综合测试套件
// 测试并发访问模式、缓存行为和同步正确性
// 验证基础操作和高级双重检查锁定实现
package mutexmap_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/mutexmap"
)

// TestMap_Getset validates lazy-loading cache behavior with double-checked locking pattern
// Tests that the first Getset computes and caches value (loaded false), the next ones retrieve cached value (loaded true)
// Ensures cached values remain stable even when compute function returns different results
//
// TestMap_Getset 验证具有双重检查锁定模式的延迟加载缓存行为
// 测试首次调用计算并缓存值（loaded 为 false），后续调用获取缓存值（loaded 为 true）
// 确保即使计算函数返回不同结果，缓存值仍保持稳定
func TestMap_Getset(t *testing.T) {
	a := mutexmap.NewMap[int, string](0)
	{
		v, loaded := a.Getset(0, func() string {
			return "abc"
		})
		require.False(t, loaded)
		require.Equal(t, "abc", v)
	}
	{
		v, loaded := a.Getset(0, func() string {
			return "xyz"
		})
		require.True(t, loaded)
		require.Equal(t, "abc", v) // the value stays when the ID exists
	}
}

// TestMap_ZeroValue checks a zero-value Map works with no New, the same as sync.Map
// TestMap_ZeroValue 检查零值的 Map 不调用 New 也能直接用，跟 sync.Map 一样
func TestMap_ZeroValue(t *testing.T) {
	var a mutexmap.Map[string, int]

	v, ok := a.Get("abc")
	require.False(t, ok)
	require.Equal(t, 0, v)

	v, loaded := a.Getset("abc", func() int { return 42 })
	require.False(t, loaded)
	require.Equal(t, 42, v)

	a.Set("xyz", 7)
	a.Delete("abc")
	require.Equal(t, 1, a.Len())
}

// TestMap_Get verifies basic Set and Get operations with thread-safe access
// Tests successful storage and retrieval of values by key
// Validates proper handling of non-existent keys returning false status
//
// TestMap_Get 验证具有线程安全访问的基础 Set 和 Get 操作
// 测试按键成功存储和检索值
// 验证不存在的键正确返回 false 状态
func TestMap_Get(t *testing.T) {
	m := mutexmap.New[int, string]()
	m.Set(1, "value1")

	// 测试正常获取
	value, ok := m.Get(1)
	require.True(t, ok)
	require.Equal(t, "value1", value)

	// 测试获取不存在的键
	_, ok = m.Get(2)
	require.False(t, ok)
}

// TestMap_Set validates thread-safe concurrent write and read operations
// Spawns 100 goroutines performing simultaneous Set operations
// Verifies data integrity after concurrent modifications complete
//
// TestMap_Set 验证线程安全的并发写入和读取操作
// 生成 100 个协程执行同步 Set 操作
// 验证并发修改完成后的数据完整性
func TestMap_Set(t *testing.T) {
	m := mutexmap.New[int, string]()
	var wg sync.WaitGroup

	// 模拟并发写入
	for i := range 100 {
		wg.Go(func() {
			m.Set(i, "v"+strconv.Itoa(i))
		})
	}

	wg.Wait()

	// 验证写入结果
	for i := range 100 {
		value, ok := m.Get(i)
		require.True(t, ok)
		require.Equal(t, "v"+strconv.Itoa(i), value)
	}
}

// TestNew checks New gives a map with no entries, at once open to use
// TestNew 检查 New 返回一个可以直接用的空 map
func TestNew(t *testing.T) {
	m := mutexmap.New[string, int]()
	require.Equal(t, 0, m.Len())

	m.Set("a", 1)
	require.Equal(t, 1, m.Len())
}

// TestNewMap checks NewMap gives a map with no entries, the size being a starting hint
// TestNewMap 检查 NewMap 返回一个空 map，size 只是初始容量提示
func TestNewMap(t *testing.T) {
	m := mutexmap.NewMap[string, int](16)
	require.Equal(t, 0, m.Len())

	m.Set("a", 1)
	value, ok := m.Get("a")
	require.True(t, ok)
	require.Equal(t, 1, value)
}

// TestMap_Delete checks Delete removes the ID, and deleting an absent ID does nothing
// TestMap_Delete 检查 Delete 删掉这个 ID，删一个不存在的 ID 什么也不做
func TestMap_Delete(t *testing.T) {
	m := mutexmap.New[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)

	m.Delete("a")
	_, ok := m.Get("a")
	require.False(t, ok)
	require.Equal(t, 1, m.Len())

	m.Delete("absent")
	require.Equal(t, 1, m.Len())
}

// TestMap_Len checks Len counts each ID once, setting an ID again keeps the count
// TestMap_Len 检查 Len 每个 ID 只算一次，同一个 ID 再设置一遍数量不变
func TestMap_Len(t *testing.T) {
	m := mutexmap.New[string, int]()
	require.Equal(t, 0, m.Len())

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("a", 3)
	require.Equal(t, 2, m.Len())
}

// TestMap_Range checks Range visits each ID, and stops once the function returns false
// TestMap_Range 检查 Range 访问每个 ID，函数返回 false 时停止
func TestMap_Range(t *testing.T) {
	m := mutexmap.New[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	visited := map[string]int{}
	m.Range(func(k string, v int) bool {
		visited[k] = v
		return true
	})
	require.Equal(t, map[string]int{"a": 1, "b": 2, "c": 3}, visited)

	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return false
	})
	require.Equal(t, 1, count)
}
