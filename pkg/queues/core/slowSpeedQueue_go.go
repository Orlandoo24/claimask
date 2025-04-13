// Package queues 提供队列和速率限制相关的工具
package queues

import (
	"errors"
	"sync"
	"time"
)

// QueueMessage 表示通用的队列消息
type QueueMessage struct {
	// Address 表示消息的目标地址
	Address string
	// Value 表示消息的交易值
	Value int
	// Data 存储任意附加数据
	Data interface{}
}

// TaskProcessor 定义消息处理器函数签名
type TaskProcessor func(sender string, senderKey string, messages []QueueMessage) error

// RateLimitedQueue 实现了一个速率限制队列
// 特点:
// 1. 支持批量和单条消息处理
// 2. 可配置的处理间隔
// 3. 地址黑名单机制防止重复提交
// 4. 线程安全的并发处理
type RateLimitedQueue struct {
	// 核心配置
	messageQueue []QueueMessage      // 消息队列
	blacklist    map[string]struct{} // 黑名单集合(使用map提高查找效率)
	processor    TaskProcessor       // 消息处理器
	sender       string              // 发送方地址
	senderKey    string              // 发送方密钥

	// 速率控制
	interval     time.Duration // 处理间隔
	valueLimit   int           // 单笔交易值限制
	batchProcess bool          // 是否批量处理

	// 并发控制
	mutex      sync.Mutex    // 互斥锁
	processing bool          // 处理状态标记
	ticker     *time.Ticker  // 定时器
	done       chan struct{} // 关闭信号
}

// NewRateLimitedQueue 创建新的速率限制队列
// processor: 消息处理函数
// sender: 发送方地址
// senderKey: 发送方密钥
// interval: 处理间隔时间
// batchProcess: 是否批量处理
// valueLimit: 单笔交易值上限
func NewRateLimitedQueue(
	processor TaskProcessor,
	sender, senderKey string,
	interval time.Duration,
	batchProcess bool,
	valueLimit int,
) *RateLimitedQueue {
	q := &RateLimitedQueue{
		processor:    processor,
		sender:       sender,
		senderKey:    senderKey,
		interval:     interval,
		batchProcess: batchProcess,
		valueLimit:   valueLimit,
		blacklist:    make(map[string]struct{}),
		done:         make(chan struct{}),
	}

	// 如果是批量处理模式, 启动定时器
	if batchProcess {
		q.ticker = time.NewTicker(interval)
		go q.processingLoop()
	}

	return q
}

// Enqueue 将消息加入队列
// 返回错误表示入队失败原因
func (q *RateLimitedQueue) Enqueue(message QueueMessage) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 检查交易值限制
	if q.valueLimit > 0 && message.Value > q.valueLimit {
		return errors.New("value exceeds limit")
	}

	// 检查地址是否在黑名单中
	if _, exists := q.blacklist[message.Address]; exists {
		return errors.New("address is blacklisted")
	}

	// 将地址加入黑名单
	q.blacklist[message.Address] = struct{}{}

	// 将消息加入队列
	q.messageQueue = append(q.messageQueue, message)

	// 如果是顺序处理模式且当前没有处理任务, 启动处理
	if !q.batchProcess && !q.processing {
		q.processing = true
		go q.processSequential()
	}

	return nil
}

// RemoveFromBlacklist 从黑名单中移除地址
func (q *RateLimitedQueue) RemoveFromBlacklist(address string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	delete(q.blacklist, address)
}

// processingLoop 批量处理循环
func (q *RateLimitedQueue) processingLoop() {
	for {
		select {
		case <-q.ticker.C:
			q.processBatch()
		case <-q.done:
			q.ticker.Stop()
			return
		}
	}
}

// processBatch 批量处理当前队列中的所有消息
func (q *RateLimitedQueue) processBatch() {
	q.mutex.Lock()
	if len(q.messageQueue) == 0 {
		q.mutex.Unlock()
		return
	}

	// 复制当前队列内容并清空队列
	messages := make([]QueueMessage, len(q.messageQueue))
	copy(messages, q.messageQueue)
	q.messageQueue = q.messageQueue[:0]
	q.mutex.Unlock()

	// 异步处理消息批次
	go func() {
		if err := q.processor(q.sender, q.senderKey, messages); err != nil {
			// 处理失败可以选择重新入队或记录日志
		}
	}()
}

// processSequential 顺序处理队列消息
func (q *RateLimitedQueue) processSequential() {
	defer func() {
		q.mutex.Lock()
		q.processing = false
		q.mutex.Unlock()
	}()

	for {
		// 获取队首消息
		q.mutex.Lock()
		if len(q.messageQueue) == 0 {
			q.mutex.Unlock()
			return
		}

		message := q.messageQueue[0]
		q.messageQueue = q.messageQueue[1:]
		q.mutex.Unlock()

		// 处理单条消息
		err := q.processor(q.sender, q.senderKey, []QueueMessage{message})
		if err != nil {
			// 这里可以实现错误处理策略
		}

		// 固定延迟
		time.Sleep(q.interval)
	}
}

// Close 关闭队列及清理资源
func (q *RateLimitedQueue) Close() {
	close(q.done)
}

// QueueStats 返回队列当前状态信息
type QueueStats struct {
	QueueLength     int
	BlacklistLength int
	IsProcessing    bool
}

// GetStats 获取当前队列状态统计
func (q *RateLimitedQueue) GetStats() QueueStats {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return QueueStats{
		QueueLength:     len(q.messageQueue),
		BlacklistLength: len(q.blacklist),
		IsProcessing:    q.processing,
	}
}
