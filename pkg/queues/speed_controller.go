package queues

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SpeedController 实现了交易速率控制功能，用于限制并发任务执行的数量和速率。
// 它结合了优先级队列、信号量和重试机制，确保高优先级任务优先处理，同时防止系统过载。
// SpeedController是线程安全的，可以被多个goroutine并发访问。
type SpeedController struct {
	mu          sync.Mutex       // 互斥锁，保证并发安全
	priorityQ   *PriorityQueue   // 存储待处理任务的优先级队列
	workerSem   chan struct{}    // 并发控制信号量，限制同时执行的任务数量
	retryPolicy RetryPolicy      // 失败任务的重试策略
	pending     map[string]*Item // 记录正在处理的任务，用于去重和状态跟踪
}

// RetryPolicy 定义了任务失败后的重试策略。
// 它包含重试次数限制和各种退避延迟参数，用于实现指数退避算法。
type RetryPolicy struct {
	// MaxRetries 指定任务最大重试次数，超过此次数的任务将被标记为永久失败
	MaxRetries int
	// BaseDelay 指定第一次重试前的等待时间
	BaseDelay time.Duration
	// MaxDelay 指定重试等待的最大时间，防止退避时间过长
	MaxDelay time.Duration
}

// NewSpeedController 创建并初始化一个速率控制器。
// 传入的config参数包含并发限制和重试策略等配置信息。
// 返回一个可立即使用的SpeedController实例。
func NewSpeedController(config Config) *SpeedController {
	sc := &SpeedController{
		priorityQ:   NewPriorityQueue(100),                     // 创建一个初始容量为100的优先级队列
		workerSem:   make(chan struct{}, config.MaxConcurrent), // 创建容量为MaxConcurrent的信号量通道
		retryPolicy: config.RetryPolicy,
		pending:     make(map[string]*Item), // 初始化待处理任务映射
	}
	return sc
}

// Enqueue 将一个任务添加到处理队列中。
// 如果任务的Key已经存在于pending map中，表示相同任务正在处理中，此时会忽略新任务。
// 任务入队后，会自动启动一个goroutine来处理队列中的任务。
// 这是一个线程安全的操作。
func (sc *SpeedController) Enqueue(item *Item) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// 任务去重检查，防止同一任务被重复处理
	if _, exists := sc.pending[item.Key]; exists {
		return
	}

	heap.Push(sc.priorityQ, item) // 将任务添加到优先级队列
	sc.pending[item.Key] = item   // 记录到待处理映射中
	go sc.processTasks()          // 启动任务处理
}

// processTasks 是一个内部方法，用于处理队列中的任务。
// 它会尝试获取信号量，一旦获取成功，就会从队列中取出优先级最高的任务进行处理。
// 如果无法立即获取信号量（已达到最大并发数），此方法会立即返回。
// 任务处理完成后会释放信号量，允许处理下一个任务。
func (sc *SpeedController) processTasks() {
	for {
		select {
		case sc.workerSem <- struct{}{}: // 尝试获取信号量
			item := sc.dequeue()
			if item == nil {
				<-sc.workerSem // 如果没有任务，释放信号量并返回
				return
			}

			go sc.executeWithRetry(item) // 异步执行任务，并在失败时进行重试
		default:
			return // 如果无法获取信号量，说明已达到最大并发数，直接返回
		}
	}
}

// dequeue 从优先级队列中取出优先级最高的任务。
// 如果队列为空，返回nil。
// 任务被取出后会从pending map中移除。
// 这是一个线程安全的操作。
func (sc *SpeedController) dequeue() *Item {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.priorityQ.Len() == 0 {
		return nil
	}

	item := heap.Pop(sc.priorityQ).(*Item)
	delete(sc.pending, item.Key) // 从待处理映射中移除
	return item
}

// executeWithRetry 执行任务，并在失败时根据RetryPolicy进行重试。
// 每次执行结束后，无论成功失败，都会释放一个信号量。
// 如果任务在最大重试次数内仍然失败，会调用handleFailure方法进行处理。
func (sc *SpeedController) executeWithRetry(item *Item) {
	defer func() { <-sc.workerSem }() // 确保在函数返回时释放信号量

	for attempt := 0; attempt <= sc.retryPolicy.MaxRetries; attempt++ {
		err := item.Handler(context.Background()) // 执行任务处理器
		if err == nil {
			return // 处理成功，直接返回
		}

		delay := sc.calculateBackoff(attempt) // 计算退避延迟时间
		time.Sleep(delay)                     // 等待后重试
	}

	// 超过最大重试次数，处理失败
	sc.handleFailure(item)
}

// calculateBackoff 根据重试次数计算退避延迟时间。
// 使用指数退避算法：delay = BaseDelay * 2^attempt，但不超过MaxDelay。
// 这种算法可以在系统负载高时降低重试频率，减轻系统压力。
func (sc *SpeedController) calculateBackoff(attempt int) time.Duration {
	delay := sc.retryPolicy.BaseDelay * time.Duration(1<<uint(attempt)) // 指数增长
	if delay > sc.retryPolicy.MaxDelay {
		return sc.retryPolicy.MaxDelay // 不超过最大延迟时间
	}
	return delay
}

// handleFailure 处理达到最大重试次数仍然失败的任务。
// 如果任务的RetryCount小于最大重试次数，会增加重试计数并提升优先级，然后重新加入队列。
// 否则，会将任务记录为永久失败。
// 这是一个线程安全的操作。
func (sc *SpeedController) handleFailure(item *Item) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if item.RetryCount < sc.retryPolicy.MaxRetries {
		item.RetryCount++
		item.Priority += 10 // 提升失败任务的优先级，使其能更快被再次处理
		heap.Push(sc.priorityQ, item)
		sc.pending[item.Key] = item
	} else {
		// 记录永久失败的任务
		logFailedTransaction(item)
	}
}

// logFailedTransaction 记录永久失败的交易信息。
// 使用zap日志库记录详细的错误信息，包括交易Key、重试次数和交易数据。
// 这些日志可用于后续的问题排查和统计分析。
func logFailedTransaction(item *Item) {
	zap.L().Error("交易处理失败，已达到最大重试次数",
		zap.String("txKey", item.Key),
		zap.Int("retryCount", item.RetryCount),
		zap.Any("value", item.Value))
}
