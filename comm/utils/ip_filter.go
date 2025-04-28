package utils

import (
	"container/list"
	"sync"
	"time"

	"go.uber.org/zap"
)

// IPFilter 使用LRU算法实现的IP地址过滤器
// 用于限制短时间内来自同一IP的频繁请求
type IPFilter struct {
	// 最大缓存IP数量
	maxEntries int
	// 访问时间窗口（秒）
	windowSeconds int
	// 窗口内最大请求次数
	maxRequests int
	// 缓存，key为IP地址，值为list.Element
	cache map[string]*list.Element
	// 双向链表，用于实现LRU
	ll *list.List
	// 互斥锁保护并发访问
	mu sync.Mutex
}

// ipEntry 表示IP访问记录的条目
type ipEntry struct {
	key       string      // IP地址
	visits    []time.Time // 访问时间历史记录
	blacklist time.Time   // 黑名单到期时间（如果被加入黑名单）
}

// NewIPFilter 创建一个新的IP过滤器
// maxEntries: 最多缓存多少个不同IP地址
// windowSeconds: 时间窗口大小，单位秒
// maxRequests: 时间窗口内允许的最大请求次数
func NewIPFilter(maxEntries, windowSeconds, maxRequests int) *IPFilter {
	return &IPFilter{
		maxEntries:    maxEntries,
		windowSeconds: windowSeconds,
		maxRequests:   maxRequests,
		cache:         make(map[string]*list.Element),
		ll:            list.New(),
	}
}

// Allow 检查IP是否允许访问
// 返回true表示允许，false表示拒绝
func (f *IPFilter) Allow(ip string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()

	// 检查IP是否在缓存中
	if element, exists := f.cache[ip]; exists {
		entry := element.Value.(*ipEntry)

		// 检查是否在黑名单中
		if entry.blacklist.After(now) {
			return false
		}

		// 更新LRU位置
		f.ll.MoveToFront(element)

		// 清理过期的访问记录
		entry.visits = filterRecentVisits(entry.visits, now, f.windowSeconds)

		// 检查访问频率
		if len(entry.visits) >= f.maxRequests {
			// 加入黑名单10分钟
			entry.blacklist = now.Add(10 * time.Minute)
			zap.L().Info("IP已被暂时限制访问",
				zap.String("ip", ip),
				zap.String("reason", "请求频率过高"))
			return false
		}

		// 记录此次访问
		entry.visits = append(entry.visits, now)
		return true
	}

	// 如果缓存已满，移除最久未使用的条目
	if f.ll.Len() >= f.maxEntries {
		oldest := f.ll.Back()
		if oldest != nil {
			f.removeElement(oldest)
		}
	}

	// 添加新的IP记录
	entry := &ipEntry{
		key:    ip,
		visits: []time.Time{now},
	}
	element := f.ll.PushFront(entry)
	f.cache[ip] = element
	return true
}

// removeElement 从缓存和链表中移除元素
func (f *IPFilter) removeElement(e *list.Element) {
	f.ll.Remove(e)
	entry := e.Value.(*ipEntry)
	delete(f.cache, entry.key)
}

// CleanBlacklist 清理已过期的黑名单IP
func (f *IPFilter) CleanBlacklist() {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	for _, element := range f.cache {
		entry := element.Value.(*ipEntry)
		if !entry.blacklist.IsZero() && entry.blacklist.Before(now) {
			entry.blacklist = time.Time{} // 重置黑名单时间
			entry.visits = []time.Time{}  // 清空访问记录
		}
	}
}

// filterRecentVisits 仅保留时间窗口内的访问记录
func filterRecentVisits(visits []time.Time, now time.Time, windowSeconds int) []time.Time {
	cutoff := now.Add(-time.Duration(windowSeconds) * time.Second)
	i := 0
	for ; i < len(visits); i++ {
		if visits[i].After(cutoff) {
			break
		}
	}

	if i == 0 {
		return visits
	}

	return visits[i:]
}

// GetIPStats 获取指定IP的统计信息
func (f *IPFilter) GetIPStats(ip string) (visits int, isBlacklisted bool, timeLeft time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	if element, exists := f.cache[ip]; exists {
		entry := element.Value.(*ipEntry)

		// 计算在窗口内的访问次数
		visits = len(filterRecentVisits(entry.visits, now, f.windowSeconds))

		// 检查是否在黑名单中
		isBlacklisted = entry.blacklist.After(now)
		if isBlacklisted {
			timeLeft = entry.blacklist.Sub(now)
		}
	}

	return
}

// RunCleaner 启动定期清理过期黑名单的goroutine
func (f *IPFilter) RunCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			f.CleanBlacklist()
		}
	}()
}

// 使用示例
// 创建过滤器: 最多缓存1000个IP，60秒内最多允许10次请求
// ipFilter := utils.NewIPFilter(1000, 60, 30)

// // 启动定期清理
// ipFilter.RunCleaner()

// // 在HTTP中间件中使用
// func IPLimitMiddleware(next http.Handler) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         ip := getClientIP(r)
//         if !ipFilter.Allow(ip) {
//             http.Error(w, "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
//             return
//         }
//         next.ServeHTTP(w, r)
//     })
// }
