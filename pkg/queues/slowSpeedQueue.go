package queues

import (
	"errors"
	"sync"
	"time"
)

// Message 表示一个需要处理的消息，包含目标地址和交易值。
// 主要用于SlowSpeedBox的批量处理场景。
type Message struct {
	// Address 表示消息的目标地址
	Address string
	// Value 表示消息的交易金额或数量
	Value int
}

// SlowSpeedBox 实现了一个批量消息处理系统，用于周期性地处理累积的消息。
// 它内部使用定时器定期处理队列中的消息，同时实现了防重复提交和地址黑名单功能。
// 该结构适用于需要批量处理且有频率限制的场景，如批量发送交易。
type SlowSpeedBox struct {
	addressQueue []Message                                            // 存储待处理的消息队列
	banAddress   []string                                             // 存储被禁止处理的地址列表
	fun          func(address, privateKey string, messages []Message) // 消息处理函数
	address      string                                               // 处理消息的账户地址
	privateKey   string                                               // 处理消息的账户私钥
	senderTicker *time.Ticker                                         // 控制消息处理频率的定时器
	banderTicker *time.Ticker                                         // 控制黑名单清理频率的定时器
	mutex        sync.Mutex                                           // 保证并发安全的互斥锁
}

// NewSlowSpeedBox 创建并初始化一个新的SlowSpeedBox实例。
// 参数fun是处理消息的函数，将在定时器触发时被调用。
// 参数address和privateKey是发送方的账户信息。
// 返回一个已启动内部定时器的SlowSpeedBox指针。
func NewSlowSpeedBox(
	fun func(address, privateKey string, messages []Message),
	address, privateKey string,
) *SlowSpeedBox {
	s := &SlowSpeedBox{
		fun:        fun,
		address:    address,
		privateKey: privateKey,
	}

	// 初始化定时器：每60秒处理一次消息队列
	s.senderTicker = time.NewTicker(60 * time.Second)
	// 初始化定时器：每24小时清理一次黑名单
	s.banderTicker = time.NewTicker(24 * time.Hour)

	// 启动后台处理goroutine
	go func() {
		for {
			select {
			case <-s.senderTicker.C:
				s.processBatch() // 定期处理批次消息
			case <-s.banderTicker.C:
				s.clearBanned() // 定期清理黑名单
			}
		}
	}()

	return s
}

// processBatch 处理当前队列中的所有消息。
// 该方法会将消息队列中的所有消息一次性提取出来，然后异步调用处理函数。
// 处理完成后，队列会被清空，为新的消息做准备。
// 这是一个线程安全的操作。
func (s *SlowSpeedBox) processBatch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.addressQueue) == 0 {
		return // 队列为空，无需处理
	}

	// 复制当前队列并清空原队列
	messages := make([]Message, len(s.addressQueue))
	copy(messages, s.addressQueue)
	s.addressQueue = s.addressQueue[:0]

	// 异步处理复制出的消息批次
	go func() {
		s.fun(s.address, s.privateKey, messages)
		// 这里可以添加日志记录或结果回调
	}()
}

// clearBanned 清空黑名单列表，允许之前被禁止的地址再次提交消息。
// 该方法通常由内部定时器自动调用，周期为24小时。
// 这是一个线程安全的操作。
func (s *SlowSpeedBox) clearBanned() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.banAddress = s.banAddress[:0] // 清空黑名单
}

// Enqueue 将一个新消息添加到处理队列中。
// 该方法会进行多项检查:
// 1. 验证消息值是否超过单笔限额(5000)
// 2. 检查目标地址是否在黑名单中
// 3. 确保同一地址不会在一个批次中重复出现
// 如果所有检查通过，则将消息添加到队列并将地址加入黑名单。
// 返回error表示添加失败的原因，nil表示添加成功。
// 这是一个线程安全的操作。
func (s *SlowSpeedBox) Enqueue(message Message) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查单笔交易金额是否超过限制
	if message.Value > 5000 {
		return errors.New("single transaction exceeds 5000 limit")
	}

	// 检查地址是否在黑名单中
	for _, addr := range s.banAddress {
		if addr == message.Address {
			return errors.New("address banned within 24 hours")
		}
	}

	// 检查地址是否已在当前批次中
	for _, msg := range s.addressQueue {
		if msg.Address == message.Address {
			return errors.New("address already in processing queue")
		}
	}

	// 将地址加入黑名单并添加消息到队列
	s.banAddress = append(s.banAddress, message.Address)
	s.addressQueue = append(s.addressQueue, message)
	return nil
}

// TaskHandler 定义了处理单个任务的函数类型。
// 该函数接收发送方地址、私钥和任务数据，返回处理结果。
type TaskHandler func(address, privateKey string, data interface{}) error

// SlowSpeedQueue 实现了一个简单的顺序任务处理队列。
// 与SlowSpeedBox不同，它不进行批处理，而是一个接一个地处理任务，
// 每个任务处理完成后会等待固定时间再处理下一个任务。
// 这种设计适用于需要严格控制处理间隔的场景。
type SlowSpeedQueue struct {
	queue      []interface{} // 待处理的任务队列
	processing bool          // 标记是否正在处理队列
	fun        TaskHandler   // 任务处理函数
	address    string        // 处理任务的账户地址
	privateKey string        // 处理任务的账户私钥
	mutex      sync.Mutex    // 保证并发安全的互斥锁
}

// NewSlowSpeedQueue 创建并初始化一个新的SlowSpeedQueue实例。
// 参数fun是处理任务的函数。
// 参数address和privateKey是发送方的账户信息。
// 返回一个初始化完成的SlowSpeedQueue指针。
func NewSlowSpeedQueue(fun TaskHandler, address, privateKey string) *SlowSpeedQueue {
	return &SlowSpeedQueue{
		fun:        fun,
		address:    address,
		privateKey: privateKey,
	}
}

// Enqueue 将一个新任务添加到处理队列中。
// 如果队列当前没有在处理任务，会自动启动处理循环。
// 这是一个线程安全的操作。
func (q *SlowSpeedQueue) Enqueue(data interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 添加任务到队列
	q.queue = append(q.queue, data)

	// 如果队列不在处理状态，启动处理循环
	if !q.processing {
		q.processing = true
		go q.processLoop()
	}
}

// processLoop 是一个内部方法，用于循环处理队列中的任务。
// 它会依次处理队列中的每个任务，每次处理后等待15秒再处理下一个任务。
// 当队列中没有更多任务时，处理循环会结束。
// 该方法会在Enqueue方法首次添加任务时被启动，
// 确保每个任务都有充分的处理时间和网络资源。
func (q *SlowSpeedQueue) processLoop() {
	defer func() {
		q.mutex.Lock()
		q.processing = false
		q.mutex.Unlock()
	}()

	for {
		q.mutex.Lock()
		if len(q.queue) == 0 {
			q.mutex.Unlock()
			return // 队列为空，结束处理循环
		}

		// 取出队首任务并从队列中移除
		task := q.queue[0]
		q.queue = q.queue[1:]
		q.mutex.Unlock()

		// 执行任务处理函数
		if err := q.fun(q.address, q.privateKey, task); err != nil {
			// 这里可以添加错误处理逻辑
		}

		// 固定延迟，控制处理速率
		time.Sleep(15 * time.Second)
	}
}
