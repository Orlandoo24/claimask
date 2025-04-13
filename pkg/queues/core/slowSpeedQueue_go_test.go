package queues

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 测试基本功能: 创建队列和检查初始状态
func TestNewRateLimitedQueue(t *testing.T) {
	processor := func(sender, senderKey string, messages []QueueMessage) error {
		return nil
	}

	// 测试批量处理模式
	batchQueue := NewRateLimitedQueue(
		processor,
		"testAddress",
		"testKey",
		100*time.Millisecond,
		true, // 批量处理
		1000,
	)
	defer batchQueue.Close()

	// 检查初始状态
	stats := batchQueue.GetStats()
	if stats.QueueLength != 0 {
		t.Errorf("Expected empty queue, got length %d", stats.QueueLength)
	}
	if stats.BlacklistLength != 0 {
		t.Errorf("Expected empty blacklist, got length %d", stats.BlacklistLength)
	}

	// 测试顺序处理模式
	seqQueue := NewRateLimitedQueue(
		processor,
		"testAddress",
		"testKey",
		100*time.Millisecond,
		false, // 顺序处理
		1000,
	)
	defer seqQueue.Close()

	// 检查初始状态
	stats = seqQueue.GetStats()
	if stats.QueueLength != 0 {
		t.Errorf("Expected empty queue, got length %d", stats.QueueLength)
	}
	if stats.BlacklistLength != 0 {
		t.Errorf("Expected empty blacklist, got length %d", stats.BlacklistLength)
	}
}

// 测试消息入队功能
func TestEnqueue(t *testing.T) {
	var processed bool
	processor := func(sender, senderKey string, messages []QueueMessage) error {
		processed = true
		return nil
	}

	queue := NewRateLimitedQueue(
		processor,
		"testAddress",
		"testKey",
		50*time.Millisecond,
		false, // 顺序处理模式便于测试
		1000,
	)
	defer queue.Close()

	// 测试正常入队
	err := queue.Enqueue(QueueMessage{
		Address: "addr1",
		Value:   100,
		Data:    "test data",
	})
	if err != nil {
		t.Errorf("Enqueue failed: %v", err)
	}

	// 检查队列状态
	stats := queue.GetStats()
	if stats.QueueLength != 1 {
		t.Errorf("Expected queue length 1, got %d", stats.QueueLength)
	}

	// 测试重复地址入队被拒绝
	err = queue.Enqueue(QueueMessage{
		Address: "addr1", // 相同地址
		Value:   200,
	})
	if err == nil {
		t.Error("Expected error for duplicate address, got nil")
	}

	// 测试超过值限制被拒绝
	err = queue.Enqueue(QueueMessage{
		Address: "addr2",
		Value:   2000, // 超过限制
	})
	if err == nil {
		t.Error("Expected error for value exceeding limit, got nil")
	}

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)
	if !processed {
		t.Error("Message not processed")
	}
}

// 测试黑名单管理功能
func TestBlacklistManagement(t *testing.T) {
	processor := func(sender, senderKey string, messages []QueueMessage) error {
		return nil
	}

	queue := NewRateLimitedQueue(
		processor,
		"testAddress",
		"testKey",
		100*time.Millisecond,
		true,
		1000,
	)
	defer queue.Close()

	// 添加消息并确认地址被加入黑名单
	err := queue.Enqueue(QueueMessage{Address: "addr1", Value: 100})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// 尝试重复添加应该失败
	err = queue.Enqueue(QueueMessage{Address: "addr1", Value: 100})
	if err == nil {
		t.Error("Expected error for blacklisted address, got nil")
	}

	// 从黑名单移除
	queue.RemoveFromBlacklist("addr1")

	// 移除后应该能再次入队
	err = queue.Enqueue(QueueMessage{Address: "addr1", Value: 100})
	if err != nil {
		t.Errorf("Enqueue after blacklist removal failed: %v", err)
	}
}

// 测试批量处理模式
func TestBatchProcessing(t *testing.T) {
	var processedCount int32
	wg := sync.WaitGroup{}
	wg.Add(1)

	processor := func(sender, senderKey string, messages []QueueMessage) error {
		atomic.AddInt32(&processedCount, int32(len(messages)))
		if len(messages) >= 3 {
			wg.Done() // 当至少处理3条消息时通知测试
		}
		return nil
	}

	queue := NewRateLimitedQueue(
		processor,
		"testAddress",
		"testKey",
		100*time.Millisecond, // 短间隔便于快速测试
		true,                 // 批量处理
		1000,
	)
	defer queue.Close()

	// 添加多条消息
	addresses := []string{"addr1", "addr2", "addr3", "addr4", "addr5"}
	for i, addr := range addresses {
		err := queue.Enqueue(QueueMessage{
			Address: addr,
			Value:   100 + i,
		})
		if err != nil {
			t.Errorf("Enqueue failed for address %s: %v", addr, err)
		}
	}

	// 等待处理完成
	if waitTimeout(&wg, 500*time.Millisecond) {
		t.Fatal("Timeout waiting for batch processing")
	}

	// 验证处理结果
	if atomic.LoadInt32(&processedCount) < 3 {
		t.Errorf("Expected at least 3 messages processed, got %d", processedCount)
	}
}

// 测试顺序处理模式
func TestSequentialProcessing(t *testing.T) {
	var processed []string
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(3) // 期望处理3条消息

	processor := func(sender, senderKey string, messages []QueueMessage) error {
		if len(messages) != 1 {
			t.Errorf("Expected single message in sequential mode, got %d", len(messages))
		}

		mu.Lock()
		processed = append(processed, messages[0].Address)
		mu.Unlock()

		wg.Done()
		return nil
	}

	queue := NewRateLimitedQueue(
		processor,
		"testAddress",
		"testKey",
		50*time.Millisecond, // 短间隔便于快速测试
		false,               // 顺序处理
		1000,
	)
	defer queue.Close()

	// 添加3条消息
	addresses := []string{"seq1", "seq2", "seq3"}
	for _, addr := range addresses {
		err := queue.Enqueue(QueueMessage{Address: addr, Value: 100})
		if err != nil {
			t.Errorf("Enqueue failed for address %s: %v", addr, err)
		}
	}

	// 等待处理完成
	if waitTimeout(&wg, 500*time.Millisecond) {
		t.Fatal("Timeout waiting for sequential processing")
	}

	// 验证处理顺序
	mu.Lock()
	defer mu.Unlock()
	if len(processed) != 3 {
		t.Errorf("Expected 3 processed messages, got %d", len(processed))
	}

	// 顺序应该保持一致
	for i, addr := range addresses {
		if i < len(processed) && processed[i] != addr {
			t.Errorf("Expected processing order %v, got %v", addresses, processed)
			break
		}
	}
}

// 性能测试: 批量处理吞吐量
func BenchmarkBatchProcessing(b *testing.B) {
	var count int32

	processor := func(sender, senderKey string, messages []QueueMessage) error {
		atomic.AddInt32(&count, int32(len(messages)))
		return nil
	}

	queue := NewRateLimitedQueue(
		processor,
		"benchAddress",
		"benchKey",
		10*time.Millisecond,
		true,  // 批量处理
		10000, // 高限制值避免拒绝
	)
	defer queue.Close()

	b.ResetTimer()

	// 生成不同地址确保不会被黑名单拒绝
	for i := 0; i < b.N; i++ {
		addr := fmt.Sprintf("bench_addr_%d", i)
		_ = queue.Enqueue(QueueMessage{
			Address: addr,
			Value:   100,
			Data:    "benchmark data",
		})
	}

	// 等待所有消息处理完成
	for {
		stats := queue.GetStats()
		if stats.QueueLength == 0 && !stats.IsProcessing {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	b.StopTimer()
	b.ReportMetric(float64(atomic.LoadInt32(&count)), "processed_msgs")
}

// 性能测试: 顺序处理延迟
func BenchmarkSequentialLatency(b *testing.B) {
	if b.N > 100 {
		b.N = 100 // 限制测试规模避免过长时间
	}

	var totalLatency time.Duration
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(b.N)

	startTimes := make(map[string]time.Time)

	processor := func(sender, senderKey string, messages []QueueMessage) error {
		addr := messages[0].Address

		mu.Lock()
		start := startTimes[addr]
		latency := time.Since(start)
		totalLatency += latency
		mu.Unlock()

		wg.Done()
		return nil
	}

	queue := NewRateLimitedQueue(
		processor,
		"benchAddress",
		"benchKey",
		5*time.Millisecond, // 更短的间隔用于基准测试
		false,              // 顺序处理
		10000,
	)
	defer queue.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		addr := fmt.Sprintf("latency_addr_%d", i)

		mu.Lock()
		startTimes[addr] = time.Now()
		mu.Unlock()

		queue.Enqueue(QueueMessage{
			Address: addr,
			Value:   100,
		})
	}

	wg.Wait()
	b.StopTimer()

	avgLatency := totalLatency / time.Duration(b.N)
	b.ReportMetric(float64(avgLatency)/float64(time.Millisecond), "avg_latency_ms")
}

// waitTimeout 等待 wg 完成，若超时则返回 true
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false // 正常完成
	case <-time.After(timeout):
		return true // 超时
	}
}
