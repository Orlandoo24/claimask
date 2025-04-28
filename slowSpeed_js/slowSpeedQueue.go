package slowspeed

import (
	"errors"
	"log"
	"sync"
	"time"
)

// 通用日志函数，替代JS中的logger
func logMessage(args ...interface{}) {
	log.Println(args...)
}

// 通用延时函数
func sleep(duration time.Duration) {
	time.Sleep(duration)
}

// 处理函数的类型定义
type ProcessFunc func(address string, privateKey string, messages []map[string]interface{})
type QueueProcessFunc func(args ...interface{}) error

// SlowSpeedBox 带速率限制和地址过滤的延时队列处理器
// 功能特点：
// 1. 每分钟批量处理一次队列
// 2. 24小时地址去重
// 3. 单次交易金额限制
type SlowSpeedBox struct {
	addressQueue []map[string]interface{} // 等待处理的地址队列
	banAddress   []string                 // 24小时内禁止重复操作的地址列表
	fun          ProcessFunc              // 依赖注入的实际业务处理函数
	address      string                   // 发送方钱包地址
	privateKey   string                   // 发送方私钥
	sender       *time.Ticker             // 定时处理器
	bander       *time.Ticker             // 清空禁止列表的定时器
	mutex        sync.Mutex               // 互斥锁保护队列操作
}

// NewSlowSpeedBox 创建一个新的SlowSpeedBox实例
func NewSlowSpeedBox(fun ProcessFunc, address, privateKey string) *SlowSpeedBox {
	box := &SlowSpeedBox{
		addressQueue: make([]map[string]interface{}, 0),
		banAddress:   make([]string, 0),
		fun:          fun,
		address:      address,
		privateKey:   privateKey,
		mutex:        sync.Mutex{},
	}

	// 定时处理器（每分钟触发）
	box.sender = time.NewTicker(60 * time.Second)
	go func() {
		for range box.sender.C {
			box.mutex.Lock()
			if len(box.addressQueue) == 0 {
				box.mutex.Unlock()
				continue
			}

			// 创建队列的副本用于处理
			queue := make([]map[string]interface{}, len(box.addressQueue))
			copy(queue, box.addressQueue)
			box.addressQueue = []map[string]interface{}{}
			box.mutex.Unlock()

			// 批量处理队列中的所有地址
			try := func() {
				defer func() {
					if r := recover(); r != nil {
						logMessage("slowSpeedBox Error", r)
					}
				}()

				box.fun(box.address, box.privateKey, queue)
				for _, group := range queue {
					logMessage("成功发送", group["amount"], "$doge给地址", group["address"])
				}
			}
			try()
		}
	}()

	// 24小时清空禁止列表的定时器
	box.bander = time.NewTicker(24 * time.Hour)
	go func() {
		for range box.bander.C {
			box.mutex.Lock()
			box.banAddress = []string{}
			box.mutex.Unlock()
		}
	}()

	return box
}

// Enqueue 将消息加入处理队列
func (b *SlowSpeedBox) Enqueue(message map[string]interface{}) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 安全校验层
	value, ok := message["value"].(float64)
	if ok && value > 5000 {
		return errors.New("失败，单次收益领取大于5000")
	}

	address, ok := message["address"].(string)
	if !ok {
		return errors.New("消息缺少有效的address字段")
	}

	// 检查是否在禁止列表中
	for _, banned := range b.banAddress {
		if banned == address {
			return errors.New("失败，地址24小时内领取过收益")
		}
	}

	// 防止重复入队检查
	for _, group := range b.addressQueue {
		if group["address"] == address {
			return errors.New("失败，地址已经在领取队列中")
		}
	}

	// 通过所有检查后加入队列
	b.banAddress = append(b.banAddress, address)
	b.addressQueue = append(b.addressQueue, message)

	return nil
}

// SlowSpeedQueue 串行化任务队列处理器
// 功能特点：
// 1. 先进先出顺序处理
// 2. 固定15秒间隔执行
// 3. 自动队列延续
type SlowSpeedQueue struct {
	queue      [][]interface{}  // 任务存储队列
	isPending  bool             // 处理状态锁
	fun        QueueProcessFunc // 依赖注入的业务函数
	address    string           // 相关地址
	privateKey string           // 相关私钥
	mutex      sync.Mutex       // 互斥锁保护队列操作
}

// NewSlowSpeedQueue 创建一个新的SlowSpeedQueue实例
func NewSlowSpeedQueue(fun QueueProcessFunc, address, privateKey string) *SlowSpeedQueue {
	return &SlowSpeedQueue{
		queue:      make([][]interface{}, 0),
		isPending:  false,
		fun:        fun,
		address:    address,
		privateKey: privateKey,
		mutex:      sync.Mutex{},
	}
}

// Enqueue 将任务加入处理队列
func (q *SlowSpeedQueue) Enqueue(args ...interface{}) {
	q.mutex.Lock()
	// 使用可变参数保持原始参数结构
	q.queue = append(q.queue, args)
	isPending := q.isPending
	q.mutex.Unlock()

	// 触发队列处理（如果当前空闲）
	if !isPending {
		go q.processQueue()
	}
}

// processQueue 递归处理队列的核心方法
func (q *SlowSpeedQueue) processQueue() {
	q.mutex.Lock()
	if len(q.queue) == 0 || q.isPending {
		q.mutex.Unlock()
		return
	}

	q.isPending = true // 上锁
	message := q.queue[0]
	q.queue = q.queue[1:] // 取出最早的任务
	q.mutex.Unlock()

	// 执行实际业务逻辑
	try := func() {
		defer func() {
			if r := recover(); r != nil {
				logMessage("处理任务时出错:", r)
			}
		}()

		err := q.fun(message...)
		if err != nil {
			logMessage("处理任务时出错:", err)
		}
	}
	try()

	// 固定间隔15秒
	sleep(15 * time.Second)

	q.mutex.Lock()
	q.isPending = false // 释放锁
	q.mutex.Unlock()

	// 递归处理下一个任务
	q.processQueue()
}
