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
// Tests that first call computes and caches value (CacheSet), subsequent calls retrieve cached value (CacheGet)
// Ensures cached values remain stable even when compute function returns different results
//
// TestMap_Getset 验证具有双重检查锁定模式的延迟加载缓存行为
// 测试首次调用计算并缓存值（CacheSet），后续调用获取缓存值（CacheGet）
// 确保即使计算函数返回不同结果，缓存值仍保持稳定
func TestMap_Getset(t *testing.T) {
	a := mutexmap.NewMap[int, string](0)
	{
		v, status := a.Getset(0, func() string {
			return "abc"
		})
		require.Equal(t, mutexmap.CacheSet, status)
		require.Equal(t, v, "abc")
	}
	{
		v, status := a.Getset(0, func() string {
			return "xyz"
		})
		require.Equal(t, mutexmap.CacheGet, status)
		require.Equal(t, v, "abc") // not change value when already exist.
	}
}

// TestMutexMap_SetGet verifies basic Set and Get operations with thread-safe access
// Tests successful storage and retrieval of values by key
// Validates proper handling of non-existent keys returning false status
//
// TestMutexMap_SetGet 验证具有线程安全访问的基础 Set 和 Get 操作
// 测试按键成功存储和检索值
// 验证不存在的键正确返回 false 状态
func TestMutexMap_SetGet(t *testing.T) {
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

// TestMutexMap_GoRun validates thread-safe concurrent write and read operations
// Spawns 100 goroutines performing simultaneous Set operations
// Verifies data integrity after concurrent modifications complete
//
// TestMutexMap_GoRun 验证线程安全的并发写入和读取操作
// 生成 100 个协程执行同步 Set 操作
// 验证并发修改完成后的数据完整性
func TestMutexMap_GoRun(t *testing.T) {
	m := mutexmap.New[int, string]()
	var wg sync.WaitGroup

	// 模拟并发写入
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Set(i, "v"+strconv.Itoa(i))
		}(i)
	}

	wg.Wait()

	// 验证写入结果
	for i := 0; i < 100; i++ {
		value, ok := m.Get(i)
		require.True(t, ok)
		require.Equal(t, "v"+strconv.Itoa(i), value)
	}
}
