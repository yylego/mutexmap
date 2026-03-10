// Package cachemap provides advanced thread-safe on-demand cache map implementation
// Enables concurrent value computation with fine-grained locking to skip redundant calculations
// Supports expensive operations through id-based locking instead of map-wide locks
// Caches both success outcomes and errors to skip repeated failed computations
//
// cachemap 包提供高级线程安全的延迟加载缓存 map 实现
// 通过细粒度锁实现并发值计算，防止重复计算
// 通过按键锁定机制而非全局 map 锁来支持耗时操作
// 缓存成功结果和错误，避免重复执行失败的计算
package cachemap

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/yylego/mutexmap/internal/utils"
)

// Map provides thread-safe on-demand cache with id-based computation synchronization
// Uses dual-scope locking: map-scope RWMutex and id-scope bottle RWMutex
// Enables expensive value computations without blocking the map as a whole
// Caches computation outcomes (both success and errors) to avoid redundant operations
//
// Map 提供线程安全的延迟加载缓存，支持按键计算同步
// 使用两级锁定：map 级别的 RWMutex 和按值瓶子级别的 RWMutex
// 允许耗时值计算而不阻塞整个 map
// 缓存计算结果（成功和错误）以防止重复操作
type Map[K comparable, V any] struct {
	mp    map[K]*valueBottle[V] // Map of value bottles with id-based synchronization // 值瓶子的 map，支持按键同步
	mutex *sync.RWMutex         // Map-wide access synchronization // 全局 map 访问同步
}

// New creates a new thread-safe on-demand cache map with default starting size of 8
// Returns a Map instance that supports expensive computation caching
// Use NewMap to set a custom starting size
//
// New 创建具有默认初始容量 8 的新线程安全延迟加载缓存 map
// 返回支持耗时计算缓存的 Map 实例
// 如需指定初始容量请使用 NewMap
func New[K comparable, V any]() *Map[K, V] {
	return NewMap[K, V](8)
}

// valueBottle encapsulates a cached value with its computation state and synchronization
// Represents a holding structure that contains the computed outcome and any problem
// Uses per-bottle locking to enable concurrent access to different ids
// The done flag shows completion of computation, in both success and failure cases
//
// valueBottle 封装缓存值及其计算状态和同步机制
// 代表一个容器，持有计算结果或错误
// 使用按瓶子锁定以允许对不同键的并发访问
// done 标志指示计算完成，无论成功还是失败
type valueBottle[V any] struct {
	mutex *sync.RWMutex // Per-bottle lock for value computation synchronization // 按瓶子锁，用于值计算同步
	done  bool          // Computation completed flag (true even if computation failed) // 计算完成标志（即使计算失败也为 true）
	res   V             // Cached result value // 缓存的结果值
	err   error         // Cached error from computation (prevents repeated failed attempts) // 计算中的缓存错误（防止重复失败尝试）
}

// NewMap creates a new thread-safe on-demand cache map with the given starting size
// The cap sets the starting size hint to optimize bottle map allocation
// Returns a Map instance that caches expensive computations
//
// NewMap 创建具有指定初始容量的新线程安全延迟加载缓存 map
// cap 参数设置底层瓶子 map 的初始大小提示
// 返回针对缓存耗时计算优化的 Map 实例
func NewMap[K comparable, V any](cap int) *Map[K, V] {
	return &Map[K, V]{
		mp:    make(map[K]*valueBottle[V], cap), // Create bottle map with the given size // 创建具有初始容量的瓶子 map
		mutex: &sync.RWMutex{},                  // Create RWMutex at map scope // 创建用于 map 级别锁定的读写锁
	}
}

// Get retrieves a value at the id, giving back value, outcome, and computation status
// Gives cached outcome when the id exists and computation finished
// Gives "not exist" message when id is absent from map
// Gives "not acted" message when id exists but computation is not done
//
// Get 根据键获取值，返回值、错误和计算状态
// 如果键存在且计算完成则返回缓存结果
// 如果在 map 中找不到键则返回 "not exist" 错误
// 如果键存在但计算尚未完成则返回 "not acted" 错误
func (a *Map[K, V]) Get(k K) (res V, err error, done bool) {
	a.mutex.RLock()         // Acquire read lock for cache lookup // 获取读锁以进行缓存查找
	defer a.mutex.RUnlock() // Release read lock when done // 完成时释放读锁
	vx, ok := a.mp[k]
	if !ok {
		return res, errors.New("not exist"), false
	}
	if !vx.done {
		return res, errors.New("not acted"), false
	}
	return vx.res, vx.err, true
}

// Set stores a value for a key with no associated error
// Creates a completed bottle with success status
// Equivalent to SetVe(k, v, nil) and SetKv(k, v)
//
// Set 为键存储一个值，不带错误信息
// 创建具有成功状态的已完成瓶子
// 等同于 SetVe(k, v, nil) 和 SetKv(k, v)
func (a *Map[K, V]) Set(k K, v V) {
	a.SetVe(k, v, nil)
}

// SetKv stores a value at the id with no associated outcome (alias of Set)
// Named to show id-value semantics
// Helps when clear intent about value-setting is needed
//
// SetKv 为键存储一个值，不带错误信息（Set 的别名）
// 明确命名以指示键值语义
// 当需要明确只存储值时很有用
func (a *Map[K, V]) SetKv(k K, v V) {
	a.SetVe(k, v, nil)
}

// SetKe stores an outcome at the id using a zero value
// Creates a completed bottle with failed status
// Helps cache failure states to skip repeated failed computations
//
// SetKe 为键存储一个错误，使用零值作为值
// 创建具有失败状态的已完成瓶子
// 对于缓存失败状态以防止重复失败计算很有用
func (a *Map[K, V]) SetKe(k K, err error) {
	a.SetVe(k, utils.Zero[V](), err)
}

// SetVe stores a value and a configurable outcome at id, overwriting existing items
// Creates a completed bottle with both the outcome and the problem state
// Most flexible Set method, supports both success (v, no-problem) and failure (v, problem) states
// The additional Set methods delegate to this one
//
// SetVe 为键存储值和可选错误，覆盖任何现有条目
// 创建具有结果和错误状态的已完成瓶子
// 最灵活的 Set 方法，支持成功（v, nil）和失败（v, err）状态
// 所有其他 Set 方法都委托给这个方法
func (a *Map[K, V]) SetVe(k K, v V, err error) {
	a.mutex.Lock()         // Acquire write lock for bottle creation // 获取写锁以创建瓶子
	defer a.mutex.Unlock() // Release write lock when done // 完成时释放写锁
	a.mp[k] = &valueBottle[V]{
		mutex: &sync.RWMutex{}, //其实当设置done以后就不会再用锁啦，但没有关系的，依然写在这里避免疑惑和出错
		done:  true,
		res:   v,
		err:   err,
	}
}

// Getset retrieves a value at id, computing with newValue when absent, then caches the outcome
// Uses sophisticated dual-scope locking to enable expensive computations without blocking the map
// Checks with read lock, then creates a bottle with write lock, then computes within bottle lock
// Multiple goroutines on the same id wait at bottle lock instead of the map lock
//
// Getset 根据键获取值，若不存在则使用 newValue 计算并缓存结果
// 使用精巧的两级锁定策略，使耗时计算不阻塞 map
// 首先用读锁检查，然后用写锁创建瓶子，最后在瓶子锁中计算
// 多个请求同一键的协程在瓶子锁上等待而非 map 锁
//
// Getset获取一个值，如果值不存在，则创建对象并设置到map中。
// 当不存在时，调用newValue创建新值并放入map作为缓存。
// 在调用newValue期间不锁定整个map，即使newValue特别慢且耗时。
// 因此当存在时不改变值，返回旧值。
func (a *Map[K, V]) Getset(k K, newValue func() (V, error)) (V, error) {
	if res, err, done := a.Get(k); done {
		return res, err
	}
	//在写锁内执行-因此在占用前有可能会执行别的操作-比如设置新值/删除值-设计时要考虑时序问题
	vb, mu := a.setBottle(k)
	//释放mp的写锁-因此在释放后有可能会执行别的操作-比如设置新值/删除值-设计时要考虑时序问题
	if mu != nil {
		defer mu.Unlock()

		func() {
			defer func() {
				if ove := recover(); ove != nil {
					if erp, ok := ove.(error); ok {
						vb.err = erp
					} else {
						vb.err = errors.Errorf("panic occurred. reason: %v", ove)
					}
				}
			}()
			vb.res, vb.err = newValue() //接着去计算数据，最终释放壳子的写锁，确保一个资源只被计算一次
		}()
		vb.done = true
		return vb.res, vb.err
	} else {
		if vb.done {
			return vb.res, vb.err
		} else {
			vb.mutex.RLock() //抢占壳子的读锁，当能抢到的时候说明壳子已经有数据，抢不到说明还没计算完
			defer vb.mutex.RUnlock()
			//当抢到读锁时，说明前面的执行已经完成（done），结果已经被赋值到壳子里，就可以使用和返回
			return vb.res, vb.err
		}
	}
}

// setBottle creates/retrieves a valueBottle at id, giving back the bottle and mutex when new
// Gives the bottle and its locked mutex if this goroutine made it (the invoking code should compute)
// Gives the bottle and no mutex when bottle existed (a different goroutine owns computation)
//
// setBottle 为键创建或获取一个 valueBottle，若为新创建则返回 bottle 及其互斥锁
// 如果此协程创建了瓶子，返回瓶子及其已锁定的互斥锁（调用方应计算）
// 如果瓶子已存在，返回瓶子和 nil 互斥锁（另一个协程拥有计算权）
func (a *Map[K, V]) setBottle(k K) (*valueBottle[V], *sync.RWMutex) {
	a.mutex.Lock()         //假如读不到数据再用写锁锁住，再去尝试读数据，假如确定还是读不到，就创建数据
	defer a.mutex.Unlock() //把最外面的大锁给释放掉，这把大锁的作用是避免重复创建空壳

	vx, exist := a.mp[k]
	if !exist {
		vx = &valueBottle[V]{ //创建个空壳数据
			mutex: &sync.RWMutex{},
			done:  false,
			res:   utils.Zero[V](), // no need to set zero value
			err:   nil,             // no need to set none value
		}
		a.mp[k] = vx
		vx.mutex.Lock() //把这个壳给锁住，再开始计算数据，这时map的锁即可释放，而元素的锁将生效以确保需要同样元素的请求都得等待后面运算结束，即等待元素的释放锁
		return vx, vx.mutex
	}

	return vx, nil
}

// Delete removes a key and its associated value bottle from the map
// Removes both completed and in-progress computations
// Safe to delete non-existent keys (no operation if key not found)
// Uses write lock to ensure exclusive access during removal
//
// Delete 从 map 中删除键及其关联的值瓶子
// 删除已完成和进行中的计算
// 删除不存在的键是安全的（如果键不存在则无操作）
// 使用写锁以确保删除期间的独占访问
func (a *Map[K, V]) Delete(k K) {
	a.mutex.Lock()         // Acquire write lock for bottle removal // 获取写锁以移除瓶子
	defer a.mutex.Unlock() // Release write lock when done // 完成时释放写锁
	delete(a.mp, k)
}

// Len returns the count of value bottles in the map
// Includes both completed and in-progress computations
// Provides a snapshot of the map size at the moment of call
// Uses read lock to allow concurrent Len operations with Get
//
// Len 返回 map 中值瓶子的数量
// 包括已完成和进行中的计算
// 提供调用时刻的 map 大小快照
// 使用读锁以允许与 Get 的并发 Len 操作
func (a *Map[K, V]) Len() int {
	a.mutex.RLock()         // Acquire read lock for size check // 获取读锁以检查大小
	defer a.mutex.RUnlock() // Release read lock when done // 完成时释放读锁
	return len(a.mp)
}

// Range iterates through each value bottle in the map, calling the function with cached state
// Provides access to value, error, and computation status for each key
// Stops iteration when callback returns false, enabling conditional processing
// Uses read lock to prevent modifications during iteration
//
// Range 遍历 map 中的每个值瓶子，使用缓存状态调用函数
// 为每个键提供对值、错误和计算状态的访问
// 当回调返回 false 时停止迭代，实现条件处理
// 使用读锁以防止迭代期间的修改
func (a *Map[K, V]) Range(run func(k K, v V, err error, done bool) bool) {
	a.mutex.RLock()         // Acquire read lock for stable iteration // 获取读锁以进行稳定迭代
	defer a.mutex.RUnlock() // Release read lock when done // 完成时释放读锁
	for k, v := range a.mp {
		if ok := run(k, v.res, v.err, v.done); !ok {
			return
		}
	}
}
