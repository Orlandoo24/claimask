package main

import (
	"container/list"
	"fmt"
)

// CacheItem 是缓存中存储的数据类型。
type CacheItem struct {
	Key   string // 键，用于标识缓存项
	Value int    // 值，缓存项存储的数据
}

// LRUCache 是一个LRU（最近最少使用）缓存结构。
type LRUCache struct {
	capacity int                      // 缓存的容量
	elements map[string]*list.Element // 存储键和对应链表元素的映射
	list     *list.List               // 双向链表，用于存储缓存项
}

// NewLRUCache 初始化一个新的LRUCache。
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		elements: make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get 从缓存中获取一个元素。
func (cache *LRUCache) Get(key string) (int, bool) {
	// 检查键是否存在
	if elem, found := cache.elements[key]; found {
		// 如果存在，将其移动到链表前端，表示最近使用过
		cache.list.MoveToFront(elem)
		// 返回找到的值和true
		return elem.Value.(*CacheItem).Value, true
	}
	// 如果不存在，返回0和false
	return 0, false
}

// Put 向缓存中添加一个新元素或更新现有元素。
func (cache *LRUCache) Put(key string, value int) {
	// 检查键是否已经存在
	if elem, found := cache.elements[key]; found {
		// 如果存在，更新值并移动到链表前端
		cache.list.MoveToFront(elem)
		elem.Value.(*CacheItem).Value = value
		return
	}

	// 如果缓存已满，移除最近最少使用的元素
	if cache.list.Len() == cache.capacity {
		oldest := cache.list.Back()
		cache.list.Remove(oldest)
		delete(cache.elements, oldest.Value.(*CacheItem).Key)
	}

	// 创建新的缓存项并添加到链表和映射中
	item := &CacheItem{Key: key, Value: value}
	elem := cache.list.PushFront(item)
	cache.elements[key] = elem
}

func main() {
	// 初始化一个容量为2的LRU缓存
	lru := NewLRUCache(2)

	// 添加元素
	lru.Put("one", 1)
	lru.Put("two", 2)
	// 获取元素并打印结果 lru.Get("one")
	fmt.Println(lru.Get("one")) // 应输出: 1, true

	// 添加新元素，导致键"two"被移除
	lru.Put("three", 3)
	// 尝试获取被移除的元素
	fmt.Println(lru.Get("two")) // 应输出: 0, false

	// 添加新元素，导致键"one"被移除
	lru.Put("four", 4)
	// 尝试获取被移除的元素
	fmt.Println(lru.Get("one")) // 应输出: 0, false
	// 获取存在的元素
	fmt.Println(lru.Get("three")) // 应输出: 3, true
	fmt.Println(lru.Get("four"))  // 应输出: 4, true
}
