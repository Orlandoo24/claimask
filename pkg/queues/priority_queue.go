// Package queues 提供了一组高性能、线程安全的队列实现，用于处理具有不同优先级的任务。
// 该包特别适用于需要异步处理、速率控制和重试机制的应用场景，如区块链交易处理。
package queues

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// Item 表示优先级队列中的一个任务项。
// 每个Item包含了执行任务所需的所有信息，包括唯一标识、数据、优先级和处理函数。
// Item实例在队列中会根据其Priority值进行排序处理。
type Item struct {
	// Key 表示任务的唯一标识符，用于重复检测和状态跟踪
	Key string
	// Value 存储任意类型的任务数据
	Value interface{}
	// Priority 定义任务的处理优先级，数值越大优先级越高
	Priority int64
	// Handler 是任务的处理函数，将在任务出队时被调用
	Handler TxHandler
	// RetryCount 记录任务已重试的次数
	RetryCount int
	// EnqueueTime 记录任务入队的时间戳，用于计算任务等待时间和超时检测
	EnqueueTime time.Time
	// index 是Item在堆中的索引，由heap包维护
	index int
}

// TxHandler 定义了处理任务的函数类型
// 函数接收一个context参数，允许上层应用进行超时控制或取消操作
// 返回error表示任务处理的结果，nil表示成功，非nil表示失败
type TxHandler func(ctx context.Context) error

// Config 定义队列系统的配置参数
type Config struct {
	// MaxConcurrent 指定最大并发处理任务数量
	MaxConcurrent int
	// RetryPolicy 定义任务失败时的重试策略
	RetryPolicy RetryPolicy
}

// PriorityQueue 实现了一个线程安全的优先级队列。
// 该队列基于Go标准库的heap接口，支持O(log n)时间复杂度的入队和出队操作。
// 队列中的元素根据优先级排序，高优先级的元素会先被处理。
type PriorityQueue struct {
	items []*Item      // 存储队列中的元素
	lock  sync.RWMutex // 用于保证线程安全的互斥锁
}

// NewPriorityQueue 创建并初始化一个具有指定初始容量的优先级队列。
// 参数capacity指定队列的初始容量，用于预分配内存以提高性能。
// 返回一个初始化完成并可以立即使用的PriorityQueue指针。
func NewPriorityQueue(capacity int) *PriorityQueue {
	pq := &PriorityQueue{
		items: make([]*Item, 0, capacity),
	}
	heap.Init(pq)
	return pq
}

// Len 返回队列中的元素数量。
// 该方法是heap.Interface接口的一部分。
// 这是一个线程安全的操作，使用读锁保护。
func (pq *PriorityQueue) Len() int {
	pq.lock.RLock()
	defer pq.lock.RUnlock()
	return len(pq.items)
}

// Less 比较两个元素的优先级，决定它们在队列中的顺序。
// 该方法是heap.Interface接口的一部分。
// 优先级算法：
// 1. 首先比较Priority字段，值越大优先级越高
// 2. 如果Priority相同，则先入队的元素优先级更高（FIFO原则）
// 这是一个线程安全的操作，使用读锁保护。
func (pq *PriorityQueue) Less(i, j int) bool {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	// 优先处理高Priority值的交易
	if pq.items[i].Priority != pq.items[j].Priority {
		return pq.items[i].Priority > pq.items[j].Priority
	}

	// 相同优先级时，先入队的优先处理
	return pq.items[i].EnqueueTime.Before(pq.items[j].EnqueueTime)
}

// Swap 交换队列中两个元素的位置。
// 该方法是heap.Interface接口的一部分。
// 交换元素时会同时更新它们的index字段，以维护堆的正确结构。
// 这是一个线程安全的操作，使用写锁保护。
func (pq *PriorityQueue) Swap(i, j int) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

// Push 将新元素添加到队列中。
// 该方法是heap.Interface接口的一部分。
// 新元素会被放置在合适的位置，以维护堆的属性。
// 同时会设置元素的EnqueueTime为当前时间，用于后续的优先级计算和超时检测。
// 这是一个线程安全的操作，使用写锁保护。
func (pq *PriorityQueue) Push(x interface{}) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	item := x.(*Item)
	item.index = len(pq.items)
	item.EnqueueTime = time.Now()
	pq.items = append(pq.items, item)
}

// Pop 从队列中移除并返回最高优先级的元素。
// 该方法是heap.Interface接口的一部分。
// 被移除的元素的index字段会被设置为-1，以标记其已不在队列中。
// 同时会清理对应位置的引用，以避免内存泄露。
// 这是一个线程安全的操作，使用写锁保护。
func (pq *PriorityQueue) Pop() interface{} {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // 避免内存泄漏
	item.index = -1 // 安全标记，表示元素不再在堆中
	pq.items = old[0 : n-1]
	return item
}

// Peek 查看队列中优先级最高的元素，但不将其从队列中移除。
// 如果队列为空，返回nil。
// 这是一个线程安全的操作，使用读锁保护。
func (pq *PriorityQueue) Peek() *Item {
	pq.lock.RLock()
	defer pq.lock.RUnlock()
	if len(pq.items) == 0 {
		return nil
	}
	return pq.items[0]
}

// UpdatePriority 更新队列中某个元素的优先级，并重新调整堆结构。
// 该方法可用于动态调整任务的优先级，例如根据等待时间或重试次数提升优先级。
// 更新后，元素会被移动到符合新优先级的位置。
// 这是一个线程安全的操作，使用写锁保护。
func (pq *PriorityQueue) UpdatePriority(item *Item, newPriority int64) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	item.Priority = newPriority
	heap.Fix(pq, item.index)
}

// CleanStaleItems 清理队列中超过指定时间的过期任务。
// 参数timeout指定超时时间，任何入队时间超过当前时间减去timeout的元素会被移除。
// 返回被移除的过期元素列表，调用方可以对这些元素进行进一步处理（如重试或记录）。
// 这是一个线程安全的操作，使用写锁保护。
func (pq *PriorityQueue) CleanStaleItems(timeout time.Duration) []*Item {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	var staleItems []*Item
	now := time.Now()

	for i := 0; i < len(pq.items); {
		if now.Sub(pq.items[i].EnqueueTime) > timeout {
			staleItems = append(staleItems, pq.items[i])
			heap.Remove(pq, i)
		} else {
			i++
		}
	}

	return staleItems
}
