// Package cachemap_test: Comprehensive test suite for advanced lazy-loading cache map
// Tests sophisticated two-level locking mechanism with per-key computation synchronization
// Validates error caching, panic recovery, concurrent access patterns, and value bottle semantics
// Ensures expensive computations execute once per key with proper synchronization
//
// cachemap_test: 高级延迟加载缓存 map 的综合测试套件
// 测试精巧的两级锁定机制和按键计算同步
// 验证错误缓存、panic 恢复、并发访问模式和值瓶子语义
// 确保耗时计算对每个键执行一次且具有正确的同步
package cachemap_test

import (
	"math/rand/v2"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/yylego/mutexmap/cachemap"
)

// TestNew validates cachemap creation with default initial size
// Tests basic set and get operations on new instance
// Verifies correct initialization and fundamental operations
//
// TestNew 验证使用默认初始大小创建 cachemap
// 测试新实例上的基础 set 和 get 操作
// 验证正确的初始化和基本操作
func TestNew(t *testing.T) {
	cache := cachemap.New[string, int]()
	require.NotNil(t, cache)
	require.Equal(t, 0, cache.Len())

	// Verify default capacity by adding elements
	const constK = "abc"
	cache.Set(constK, 42)
	res, err, done := cache.Get(constK)
	t.Log(res, err)
	require.True(t, done)
	require.NoError(t, err)
	require.Equal(t, 42, res)
}

// TestNewMap validates cacheMap creation with explicit capacity and Getset lazy-loading behavior
// Tests that compute function results (both success and errors) are cached and reused
// Verifies cache consistency across multiple Getset and Get calls with random value generation
//
// TestNewMap 验证使用显式容量创建 cachemap 和 Getset 延迟加载行为
// 测试计算函数结果（成功和错误）被缓存和重用
// 验证使用随机值生成的多个 Getset 和 Get 调用中的缓存一致性
func TestNewMap(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	calc := func() (int, error) {
		num := rand.IntN(100)
		if num < 50 {
			return -1, errors.New("wrong")
		}
		return num, nil
	}
	for idx := 0; idx < 10; idx++ {
		t.Log("(", idx, ")")
		const constK = "abc"
		res1, err1 := cache.Getset(constK, calc)
		t.Log(res1, err1)
		res2, err2 := cache.Getset(constK, calc)
		t.Log(res2, err2)

		if err1 != nil || err2 != nil {
			require.ErrorIs(t, err1, err2)
		}
		require.Equal(t, res1, res2)

		res3, err3, done := cache.Get(constK)
		require.True(t, done)
		if err2 != nil || err3 != nil {
			require.ErrorIs(t, err2, err3)
		}
		require.Equal(t, res2, res3)

		cache.Delete(constK)

		t.Log(cache.Len())
	}
}

// TestGet validates direct Get operation behavior with and without cached values
// Tests proper handling of non-existent keys returning error "not exist"
// Verifies successful retrieval of values after Set operation
//
// TestGet 验证有无缓存值时的直接 Get 操作行为
// 测试不存在的键正确返回错误 "not exist"
// 验证 Set 操作后成功检索值
func TestGet(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	const constK = "abc"
	res, err, done := cache.Get(constK)
	t.Log(res, err)
	require.False(t, done)
	require.Error(t, err)
	require.Equal(t, "not exist", err.Error())
	require.Equal(t, 0, res)

	cache.Set(constK, 42)
	res, err, done = cache.Get(constK)
	t.Log(res, err)
	require.True(t, done)
	require.NoError(t, err)
	require.Equal(t, 42, res)
}

// TestSet validates Set operation with value caching semantics
// Tests that Set operation prevents compute function execution in subsequent Getset calls
// Ensures explicitly set values take precedence over lazy computation
//
// TestSet 验证具有值缓存语义的 Set 操作
// 测试 Set 操作防止在后续 Getset 调用中执行计算函数
// 确保显式设置的值优先于延迟计算
func TestSet(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	calc := func() (int, error) {
		num := rand.IntN(100)
		if num < 50 {
			return -1, errors.New("wrong")
		}
		return num, nil
	}
	for idx := 0; idx < 10; idx++ {
		t.Log("(", idx, ")")
		const constK = "abc"
		var res0 = rand.IntN(100)
		cache.Set(constK, res0)

		res1, err1, done := cache.Get(constK)
		t.Log(res1, err1)
		require.True(t, done)
		require.NoError(t, err1)
		require.Equal(t, res0, res1)

		res2, err2 := cache.Getset(constK, calc)
		t.Log(res2, err2)
		require.NoError(t, err2)
		require.Equal(t, res0, res2)

		cache.Delete(constK)
		t.Log(cache.Len())
	}
}

// TestSetKv validates SetKv operation as alias for Set with value-focused semantics
// Tests explicit value assignment bypassing compute function in Getset operations
// Verifies value persistence across repeated operations with cache cleanup
//
// TestSetKv 验证 SetKv 操作作为具有值聚焦语义的 Set 别名
// 测试显式值分配在 Getset 操作中绕过计算函数
// 验证通过缓存清理实现重复操作中的值持久性
func TestSetKv(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	calc := func() (int, error) {
		num := rand.IntN(100)
		if num < 50 {
			return -1, errors.New("wrong")
		}
		return num, nil
	}
	for idx := 0; idx < 10; idx++ {
		t.Log("(", idx, ")")
		const constK = "abc"
		var res0 = rand.IntN(100)
		cache.SetKv(constK, res0)

		res1, err1, done := cache.Get(constK)
		t.Log(res1, err1)
		require.True(t, done)
		require.NoError(t, err1)
		require.Equal(t, res0, res1)

		res2, err2 := cache.Getset(constK, calc)
		t.Log(res2, err2)
		require.NoError(t, err2)
		require.Equal(t, res0, res2)

		cache.Delete(constK)
		t.Log(cache.Len())
	}
}

// TestSetKe validates SetKe operation caching errors with zero values
// Tests that cached errors persist and prevent compute function execution
// Verifies error-focused caching semantics across Getset and Get operations
//
// TestSetKe 验证 SetKe 操作缓存错误和零值
// 测试缓存的错误持续存在并防止执行计算函数
// 验证跨 Getset 和 Get 操作的错误聚焦缓存语义
func TestSetKe(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	calc := func() (int, error) {
		num := rand.IntN(100)
		if num < 50 {
			return -1, errors.New("wrong")
		}
		return num, nil
	}
	for idx := 0; idx < 10; idx++ {
		t.Log("(", idx, ")")
		const constK = "abc"
		err0 := errors.New("wrong")
		cache.SetKe(constK, err0)

		res1, err1, done := cache.Get(constK)
		t.Log(res1, err1)
		require.True(t, done)
		require.ErrorIs(t, err0, err1)
		require.Equal(t, 0, res1) // Zero value for int

		res2, err2 := cache.Getset(constK, calc)
		t.Log(res2, err2)
		require.ErrorIs(t, err0, err2)
		require.Equal(t, 0, res2)

		cache.Delete(constK)
		t.Log(cache.Len())
	}
}

// TestSetVe validates SetVe operation caching both values and errors atomically
// Tests that cached value-error pairs persist across multiple Getset calls
// Ensures atomic caching of computation results including success and failure states
//
// TestSetVe 验证 SetVe 操作原子性缓存值和错误
// 测试缓存的值-错误对在多个 Getset 调用中持续存在
// 确保计算结果的原子性缓存，包括成功和失败状态
func TestSetVe(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	calc := func() (int, error) {
		num := rand.IntN(100)
		if num < 50 {
			return -1, errors.New("wrong")
		}
		return num, nil
	}
	for idx := 0; idx < 10; idx++ {
		t.Log("(", idx, ")")
		const constK = "abc"
		var res0 = rand.IntN(100)
		var err0 error
		if res0 < 50 {
			res0 = -1
			err0 = errors.New("wrong")
		}
		cache.SetVe(constK, res0, err0)

		res1, err1 := cache.Getset(constK, calc)
		t.Log(res1, err1)

		if err0 != nil || err1 != nil {
			t.Log(err0)
			t.Log(err1)
			require.ErrorIs(t, err0, err1)
		}
		require.Equal(t, res0, res1)

		res2, err2 := cache.Getset(constK, calc)
		t.Log(res2, err2)

		if err1 != nil || err2 != nil {
			require.ErrorIs(t, err1, err2)
		}
		require.Equal(t, res1, res2)

		res3, err3, done := cache.Get(constK)
		require.True(t, done)
		if err2 != nil || err3 != nil {
			require.ErrorIs(t, err2, err3)
		}
		require.Equal(t, res2, res3)

		cache.Delete(constK)

		t.Log(cache.Len())
	}
}

// TestDelete validates Delete operation removing cached entries completely
// Tests that deleted keys return "not exist" error on subsequent Get calls
// Ensures proper cleanup of value bottles from map structure
//
// TestDelete 验证 Delete 操作完全删除缓存条目
// 测试删除的键在后续 Get 调用中返回 "not exist" 错误
// 确保从 map 结构中正确清理值瓶子
func TestDelete(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	for idx := 0; idx < 10; idx++ {
		t.Log("(", idx, ")")
		const constK = "abc"
		var res0 = rand.IntN(100)
		cache.Set(constK, res0)

		res1, err1, done := cache.Get(constK)
		t.Log(res1, err1)
		require.True(t, done)
		require.NoError(t, err1)
		require.Equal(t, res0, res1)

		cache.Delete(constK)

		res2, err2, done2 := cache.Get(constK)
		t.Log(res2, err2)
		require.False(t, done2)
		require.Error(t, err2)
		require.Equal(t, "not exist", err2.Error())
		require.Equal(t, 0, res2)

		t.Log(cache.Len())
	}
}

// TestLen validates Len operation tracking map size across modifications
// Tests size increments on Set operations and decrements on Delete operations
// Verifies accurate counting of cached value bottles in map
//
// TestLen 验证 Len 操作跟踪修改过程中的 map 大小
// 测试 Set 操作的大小增量和 Delete 操作的减量
// 验证 map 中缓存值瓶子的准确计数
func TestLen(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	for idx := 0; idx < 10; idx++ {
		t.Log("(", idx, ")")
		key := "key" + string(rune('A'+idx))
		var res0 = rand.IntN(100)
		cache.Set(key, res0)

		t.Log("Len:", cache.Len())
		require.Equal(t, idx+1, cache.Len())

		res1, err1, done := cache.Get(key)
		t.Log(res1, err1)
		require.True(t, done)
		require.NoError(t, err1)
		require.Equal(t, res0, res1)
	}

	for idx := 0; idx < 10; idx++ {
		key := "key" + string(rune('A'+idx))
		cache.Delete(key)
		t.Log("Len after delete:", cache.Len())
		require.Equal(t, 9-idx, cache.Len())
	}
}

// TestRange validates Range operation iterating over all cached entries
// Tests callback execution with cached values, errors, and completion status
// Verifies support for both complete iteration and advanced termination
//
// TestRange 验证 Range 操作遍历所有缓存条目
// 测试使用缓存值、错误和完成状态执行回调
// 验证对完整迭代和提前终止的支持
func TestRange(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	// Populate map with random values
	expected := make(map[string]int)
	for idx := 0; idx < 10; idx++ {
		key := "key" + string(rune('A'+idx))
		value := rand.IntN(100)
		cache.Set(key, value)
		expected[key] = value
	}

	// Test Range
	collected := make(map[string]int)
	cache.Range(func(k string, v int, err error, done bool) bool {
		t.Log("Range:", k, v, err, done)
		require.True(t, done)
		require.NoError(t, err)
		collected[k] = v
		return true
	})

	// Verify collected values
	for k, v := range expected {
		v2, ok := collected[k]
		require.True(t, ok)
		require.Equal(t, v, v2)
	}
	require.Equal(t, len(expected), len(collected))

	// Test Range with early termination
	count := 0
	cache.Range(func(k string, v int, err error, done bool) bool {
		count++
		return count < 5 // Stop after 5 iterations
	})
	require.Equal(t, 5, count)
}
