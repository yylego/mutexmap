// Package mutexmap provides thread-safe map implementations with fine-grained locking
// Supports concurrent read and write operations through sync.RWMutex
// Offers a basic thread-safe map here, plus an on-demand cache map in the cachemap package
// Enables safe concurrent access without outside synchronization
//
// mutexmap 包提供具有细粒度锁的线程安全 map 实现
// 通过 sync.RWMutex 支持并发读写操作
// 本包提供基础线程安全 map，cachemap 包另提供按需缓存 map
// 无需外部同步即可实现安全的并发访问
package mutexmap

import (
	"sync"
)

// Map provides a thread-safe map implementation using sync.RWMutex
// Supports generic keys (comparable) and values (any)
// Offers concurrent reads and exclusive writes
// Implements Getset method with double-checked locking pattern
// The zero value works with no New, the same as sync.Map, New and NewMap just set the starting size
// Do not duplicate a Map once used, it holds a mutex
//
// Map 提供使用 sync.RWMutex 的线程安全 map 实现
// 支持泛型键（可比较）和泛型值（任意）
// 提供并发读操作和独占写操作
// 实现具有双重检查锁定模式的 Getset 方法
// 零值可以直接使用，跟 sync.Map 一样，New 和 NewMap 只是设置初始容量
// 用过之后不要复制 Map，它里面有锁
type Map[K comparable, V any] struct {
	mp    map[K]V      // The underlying map storage, created on the first write when nil // 底层 map 存储，为 nil 时在第一次写入时创建
	mutex sync.RWMutex // RWMutex to manage concurrent access // 用于并发访问控制的读写锁
}

// New creates a new thread-safe map with default starting size of 8
// Returns a Map instance that supports concurrent operations
// Use NewMap to set a custom starting size
//
// New 创建具有默认初始容量 8 的新线程安全 map
// 返回支持并发操作的 Map 实例
// 如需指定初始容量请使用 NewMap
func New[K comparable, V any]() *Map[K, V] {
	return NewMap[K, V](8)
}

// NewMap creates a new thread-safe map with the given starting size
// The size sets the starting size hint to optimize allocation
// Returns a Map instance sized to fit the expected element count
//
// NewMap 创建具有指定初始容量的新线程安全 map
// size 参数设置底层 map 的初始大小提示
// 返回针对预期元素数量优化的 Map 实例
func NewMap[K comparable, V any](size int) *Map[K, V] {
	return &Map[K, V]{
		mp: make(map[K]V, size), // Create map with the given size // 创建具有初始容量的 map
	}
}

// Get retrieves the value associated with a given id
// Returns the value and true when found, zero value and false when absent
// Uses read lock to enable concurrent Get operations
// Thread-safe and allows multiple concurrent Get operations
//
// Get 获取与指定键关联的值
// 如果键存在则返回值和 true，否则返回零值和 false
// 使用读锁以允许并发 Get 操作
// 与其他 Get 操作的并发读取是线程安全的
func (a *Map[K, V]) Get(k K) (V, bool) {
	a.mutex.RLock()         // Acquire read lock for concurrent reads // 获取读锁以进行并发读取
	defer a.mutex.RUnlock() // Release read lock when done // 完成时释放读锁
	v, ok := a.mp[k]        // An absent ID gives the zero value and false // 键不存在时得到零值和 false
	return v, ok
}

// Set inserts a new mapping and updates when the id exists
// Replaces existing value without checking its previous state
// Uses write lock to ensure exclusive access to modifications
// Blocks any concurrent operations (Get, Set, Delete, Range) when executing
//
// Set 插入新的键值对或更新现有键的值
// 替换现有值而不检查先前状态
// 使用写锁以确保修改期间的独占访问
// 执行期间阻塞所有其他操作（Get、Set、Delete、Range）
func (a *Map[K, V]) Set(k K, v V) {
	a.mutex.Lock()         // Acquire write lock for exclusive access // 获取写锁以进行独占访问
	defer a.mutex.Unlock() // Release write lock when done // 完成时释放写锁
	a.ensureMap()
	a.mp[k] = v
}

// Delete removes the mapping from the map
// Safe to use even when the id does not exist (becomes a no-op)
// Uses write lock to ensure exclusive access when deleting
// Blocks any concurrent operations when executing
//
// Delete 从 map 中删除键值对
// 如果键不存在则无操作（删除不存在的键是安全的）
// 使用写锁以确保删除期间的独占访问
// 执行期间阻塞所有其他操作
func (a *Map[K, V]) Delete(k K) {
	a.mutex.Lock()         // Acquire write lock for exclusive access // 获取写锁以进行独占访问
	defer a.mutex.Unlock() // Release write lock when done // 完成时释放写锁
	delete(a.mp, k)
}

// Len returns the count of mappings in the map
// Provides a snapshot at the invocation time
// Uses read lock to enable concurrent Len operations with Get
// Note: the count can change at once because of concurrent modifications
//
// Len 返回 map 中键值对的数量
// 提供调用时刻的 map 大小快照
// 使用读锁以允许与 Get 的并发 Len 操作
// 由于并发修改，返回后计数可能立即改变
func (a *Map[K, V]) Len() int {
	a.mutex.RLock()         // Acquire read lock for concurrent reads // 获取读锁以进行并发读取
	defer a.mutex.RUnlock() // Release read lock when done // 完成时释放读锁
	return len(a.mp)
}

// Range iterates through each mapping, calling the given function on it
// Processes mappings in sequence, stopping when the callback returns false
// Uses read lock to block modifications but enables multiple Range executions
// Stops iteration when callback returns false, which supports selective processing
//
// Range 遍历 map 中的每个键值对，应用给定的函数
// 对每个条目调用函数，直到处理完所有条目或函数返回 false
// 使用读锁以防止迭代期间的修改，但允许并发 Range 调用
// 当回调返回 false 时停止迭代，实现条件处理
func (a *Map[K, V]) Range(run func(k K, v V) bool) {
	a.mutex.RLock()         // Acquire read lock for stable iteration // 获取读锁以进行稳定迭代
	defer a.mutex.RUnlock() // Release read lock when done // 完成时释放读锁
	for k, v := range a.mp {
		if !run(k, v) { // Stop iteration when callback returns false // 当回调返回 false 时停止迭代
			return
		}
	}
}

// Getset retrieves the value at the id, computes and stores it when absent
// Uses double-checked locking pattern to cut down lock contention and skip redundant calculations
// loaded is true when the value was in the map, false when this Getset computed it, the same as sync.Map.LoadOrStore
//
// Getset 获取与键关联的值，如果键不存在，则计算并存储新值
// 使用双重检查锁定模式以减少锁竞争并避免重复计算
// 值本来就在 map 里时 loaded 为 true，由这次调用计算出来时为 false，跟 sync.Map.LoadOrStore 一样
func (a *Map[K, V]) Getset(k K, calculate func() V) (v V, loaded bool) {
	if v, ok := a.Get(k); ok { // Fast path: read with shared lock // 快速路径：使用共享锁读取
		return v, true // Cached value found, return it // 找到缓存值，返回它
	}
	// 读锁释放，启动写锁，但假设有两个线程同时读不到，就都会同时占用写锁。
	a.mutex.Lock()
	defer a.mutex.Unlock()
	// 增加读锁以后二次确认内容是否在 map 里面，这样第二次占用写锁的线程就不会创建新对象。
	if v, ok := a.mp[k]; ok { // Second check: a concurrent goroutine might have set it // 二次检查：另一个协程可能已设置
		return v, true // Value now exists, return it // 值现已存在，返回它
	}
	// 当内容确实不在 map 里时，即首次占用写锁时，这才创建新对象，设置到 map 里。
	v = calculate() // Compute new value within write lock // 在写锁内计算新值
	// This function might be expensive. If this is a concern, use cachemap package instead.
	// 这个函数可能耗时较长，如果对此介意，可以使用 cachemap 包。
	a.ensureMap()
	a.mp[k] = v
	return v, false // New value computed and cached // 新值已计算并缓存
}

// ensureMap creates the underlying map when nil, which happens with the zero value, hold the write lock to use it
// ensureMap 在底层 map 为 nil 时创建它，零值的 Map 就是这种情况，调用时要持有写锁
func (a *Map[K, V]) ensureMap() {
	if a.mp == nil {
		a.mp = make(map[K]V)
	}
}
