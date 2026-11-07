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
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	res, ok, err := cache.Get(constK)
	t.Log(res, err)
	require.True(t, ok)
	require.NoError(t, err)
	require.Equal(t, 42, res)
}

// TestMap_ZeroValue checks a zero-value Map works with no New, the same as sync.Map
// TestMap_ZeroValue 检查零值的 Map 不调用 New 也能直接用，跟 sync.Map 一样
func TestMap_ZeroValue(t *testing.T) {
	var cache cachemap.Map[string, int]

	res, ok, err := cache.Get("abc")
	require.False(t, ok)
	require.NoError(t, err)
	require.Equal(t, 0, res)

	res, err = cache.Getset("abc", func() (int, error) { return 42, nil })
	require.NoError(t, err)
	require.Equal(t, 42, res)

	cache.Set("xyz", 7)
	require.Equal(t, 2, cache.Len())
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
	for idx := range 10 {
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

		res3, ok, err3 := cache.Get(constK)
		require.True(t, ok)
		if err2 != nil || err3 != nil {
			require.ErrorIs(t, err2, err3)
		}
		require.Equal(t, res2, res3)

		cache.Delete(constK)

		t.Log(cache.Len())
	}
}

// TestMap_Get checks Get with and without cached values
// Tests that an absent id gives ok false and a nil error, a miss is not an error
// Verifies successful retrieval of values after Set operation
//
// TestMap_Get 验证有无缓存值时的直接 Get 操作行为
// 测试不存在的键返回 ok 为 false、错误为 nil，「没找到」不算错误
// 验证 Set 操作后成功检索值
func TestMap_Get(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	const constK = "abc"
	res, ok, err := cache.Get(constK)
	t.Log(res, err)
	require.False(t, ok)
	require.NoError(t, err)
	require.Equal(t, 0, res)

	cache.Set(constK, 42)
	res, ok, err = cache.Get(constK)
	t.Log(res, err)
	require.True(t, ok)
	require.NoError(t, err)
	require.Equal(t, 42, res)
}

// TestMap_Set validates Set operation with value caching semantics
// Tests that Set operation prevents compute function execution in subsequent Getset calls
// Ensures explicitly set values take precedence over lazy computation
//
// TestMap_Set 验证具有值缓存语义的 Set 操作
// 测试 Set 操作防止在后续 Getset 调用中执行计算函数
// 确保显式设置的值优先于延迟计算
func TestMap_Set(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	calc := func() (int, error) {
		num := rand.IntN(100)
		if num < 50 {
			return -1, errors.New("wrong")
		}
		return num, nil
	}
	for idx := range 10 {
		t.Log("(", idx, ")")
		const constK = "abc"
		var res0 = rand.IntN(100)
		cache.Set(constK, res0)

		res1, ok, err1 := cache.Get(constK)
		t.Log(res1, err1)
		require.True(t, ok)
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

// TestMap_SetVe checks SetVe caches both values and errors at once
// Tests that cached value-error pairs persist across multiple Getset calls
// Ensures atomic caching of computation results including success and failure states
//
// TestMap_SetVe 验证 SetVe 操作原子性缓存值和错误
// 测试缓存的值-错误对在多个 Getset 调用中持续存在
// 确保计算结果的原子性缓存，包括成功和失败状态
func TestMap_SetVe(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	calc := func() (int, error) {
		num := rand.IntN(100)
		if num < 50 {
			return -1, errors.New("wrong")
		}
		return num, nil
	}
	for idx := range 10 {
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

		res3, ok, err3 := cache.Get(constK)
		require.True(t, ok)
		if err2 != nil || err3 != nil {
			require.ErrorIs(t, err2, err3)
		}
		require.Equal(t, res2, res3)

		cache.Delete(constK)

		t.Log(cache.Len())
	}
}

// TestMap_Delete validates Delete removes the cached entries
// Tests that deleted keys give ok false and a nil error on subsequent Get calls
// Ensures proper cleanup of value bottles from map structure
//
// TestMap_Delete 验证 Delete 操作完全删除缓存条目
// 测试删除的键在后续 Get 调用中返回 ok 为 false、错误为 nil
// 确保从 map 结构中正确清理值瓶子
func TestMap_Delete(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	for idx := range 10 {
		t.Log("(", idx, ")")
		const constK = "abc"
		var res0 = rand.IntN(100)
		cache.Set(constK, res0)

		res1, ok1, err1 := cache.Get(constK)
		t.Log(res1, err1)
		require.True(t, ok1)
		require.NoError(t, err1)
		require.Equal(t, res0, res1)

		cache.Delete(constK)

		res2, ok2, err2 := cache.Get(constK)
		t.Log(res2, err2)
		require.False(t, ok2)
		require.NoError(t, err2)
		require.Equal(t, 0, res2)

		t.Log(cache.Len())
	}
}

// TestMap_Len validates Len operation tracking map size across modifications
// Tests size increments on Set operations and decrements on Delete operations
// Verifies accurate counting of cached value bottles in map
//
// TestMap_Len 验证 Len 操作跟踪修改过程中的 map 大小
// 测试 Set 操作的大小增量和 Delete 操作的减量
// 验证 map 中缓存值瓶子的准确计数
func TestMap_Len(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	for idx := range 10 {
		t.Log("(", idx, ")")
		id := "id" + string(rune('A'+idx))
		var res0 = rand.IntN(100)
		cache.Set(id, res0)

		t.Log("Len:", cache.Len())
		require.Equal(t, idx+1, cache.Len())

		res1, ok, err1 := cache.Get(id)
		t.Log(res1, err1)
		require.True(t, ok)
		require.NoError(t, err1)
		require.Equal(t, res0, res1)
	}

	for idx := range 10 {
		id := "id" + string(rune('A'+idx))
		cache.Delete(id)
		t.Log("Len after delete:", cache.Len())
		require.Equal(t, 9-idx, cache.Len())
	}
}

// TestMap_Range validates Range visits the cached entries
// Tests callback execution with cached values, errors, and completion status
// Verifies support for both complete iteration and advanced termination
//
// TestMap_Range 验证 Range 操作遍历所有缓存条目
// 测试使用缓存值、错误和完成状态执行回调
// 验证对完整迭代和提前终止的支持
func TestMap_Range(t *testing.T) {
	cache := cachemap.NewMap[string, int](100)

	// Populate map with random values
	expected := make(map[string]int)
	for idx := range 10 {
		key := "key" + string(rune('A'+idx))
		value := rand.IntN(100)
		cache.Set(key, value)
		expected[key] = value
	}

	// Test Range
	collected := make(map[string]int)
	cache.Range(func(k string, v int, ok bool, err error) bool {
		t.Log("Range:", k, v, ok, err)
		require.True(t, ok)
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
	cache.Range(func(k string, v int, ok bool, err error) bool {
		count++
		return count < 5 // Stop after 5 iterations
	})
	require.Equal(t, 5, count)
}

// TestMap_Getset checks goroutines asking the same id at once get the same value, computed once,
// while Get and Range read the map during the computation (run with -race to see the data race this guards against)
//
// TestMap_Getset 检查多个协程同时取同一个键时拿到同一个值，且只计算一次，
// 计算进行时 Get 和 Range 同时在读 map（用 -race 运行，能看出它防的数据竞争）
func TestMap_Getset(t *testing.T) {
	cache := cachemap.New[string, int]()
	var calls atomic.Int32

	var wg sync.WaitGroup
	results := make([]int, 10)
	for idx := range results {
		wg.Go(func() {
			res, err := cache.Getset("slow", func() (int, error) {
				calls.Add(1)
				time.Sleep(50 * time.Millisecond)
				return 42, nil
			})
			require.NoError(t, err)
			results[idx] = res
		})
	}
	for range 10 {
		wg.Go(func() {
			for range 100 {
				// Once done, the value read is the computed one
				// 一旦算完，读到的就是计算出的值
				res, ok, err := cache.Get("slow")
				if ok {
					require.NoError(t, err)
					require.Equal(t, 42, res)
				}
				cache.Range(func(k string, v int, ok bool, err error) bool {
					if ok {
						require.NoError(t, err)
						require.Equal(t, 42, v)
					}
					return true
				})
			}
		})
	}
	wg.Wait()

	require.Equal(t, int32(1), calls.Load())
	require.Equal(t, []int{42, 42, 42, 42, 42, 42, 42, 42, 42, 42}, results)
}
