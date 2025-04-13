package slowSpeed_js

import (
	"errors"
	"sync"
	"time"
)

// 以下为 slowSpeedBox 实现

type Message struct {
	Address string
	Value   int
}

type SlowSpeedBox struct {
	addressQueue []Message
	banAddress   []string
	fun          func(address, privateKey string, messages []Message)
	address      string
	privateKey   string
	senderTicker *time.Ticker
	banderTicker *time.Ticker
	mutex        sync.Mutex
}

func NewSlowSpeedBox(
	fun func(address, privateKey string, messages []Message),
	address, privateKey string,
) *SlowSpeedBox {
	s := &SlowSpeedBox{
		fun:        fun,
		address:    address,
		privateKey: privateKey,
	}

	s.senderTicker = time.NewTicker(60 * time.Second)
	s.banderTicker = time.NewTicker(24 * time.Hour)

	go func() {
		for {
			select {
			case <-s.senderTicker.C:
				s.processBatch()
			case <-s.banderTicker.C:
				s.clearBanned()
			}
		}
	}()

	return s
}

func (s *SlowSpeedBox) processBatch() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.addressQueue) == 0 {
		return
	}

	// 复制当前队列并清空
	messages := make([]Message, len(s.addressQueue))
	copy(messages, s.addressQueue)
	s.addressQueue = s.addressQueue[:0]

	// 异步处理
	go func() {
		s.fun(s.address, s.privateKey, messages)
		// 这里可以添加日志记录逻辑
	}()
}

func (s *SlowSpeedBox) clearBanned() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.banAddress = s.banAddress[:0]
}

func (s *SlowSpeedBox) Enqueue(message Message) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if message.Value > 5000 {
		return errors.New("single transaction exceeds 5000 limit")
	}

	for _, addr := range s.banAddress {
		if addr == message.Address {
			return errors.New("address banned within 24 hours")
		}
	}

	for _, msg := range s.addressQueue {
		if msg.Address == message.Address {
			return errors.New("address already in processing queue")
		}
	}

	s.banAddress = append(s.banAddress, message.Address)
	s.addressQueue = append(s.addressQueue, message)
	return nil
}

// 以下为 slowSpeedQueue 实现

type TaskHandler func(address, privateKey string, data interface{}) error

type SlowSpeedQueue struct {
	queue      []interface{}
	processing bool
	fun        TaskHandler
	address    string
	privateKey string
	mutex      sync.Mutex
}

func NewSlowSpeedQueue(fun TaskHandler, address, privateKey string) *SlowSpeedQueue {
	return &SlowSpeedQueue{
		fun:        fun,
		address:    address,
		privateKey: privateKey,
	}
}

func (q *SlowSpeedQueue) Enqueue(data interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.queue = append(q.queue, data)

	if !q.processing {
		q.processing = true
		go q.processLoop()
	}
}

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
			return
		}

		task := q.queue[0]
		q.queue = q.queue[1:]
		q.mutex.Unlock()

		// 执行任务
		if err := q.fun(q.address, q.privateKey, task); err != nil {
			// 处理错误日志
		}

		// 固定间隔
		time.Sleep(15 * time.Second)
	}
}
