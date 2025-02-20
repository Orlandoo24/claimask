package main

import (
	"container/list"
	"fmt"
)

/*
*

	LRUCache 是一个LRU缓存，它支持以下操作：
	- Get 从缓存中获取一个元素
	- Put 向缓存中添加一个新元素或更新现有元素

*
*/
type LRUCache struct {

	// 缓存的容量
	capacity int

	// 哈希表，用于快速查找元素
	cache map[int]*list.Element

	// 双向链表，用于维护元素的LRU顺序
	queue *list.List
}

// entry 是双向链表节点的数据类型
type entry struct {
	key   int // 键
	value int // 值
}

// Constructor 初始化一个新的LRUCache
func Constructor(capacity int) LRUCache {
	return LRUCache{
		capacity: capacity,                    // 设置缓存容量
		cache:    make(map[int]*list.Element), // 初始化哈希表
		queue:    list.New(),                  // 初始化双向链表
	}
}

// Get 从缓存中获取一个元素的值
func (this *LRUCache) Get(key int) int {
	if elem, found := this.cache[key]; found { // 如果元素在缓存中
		this.queue.MoveToFront(elem)     // 将元素移动到双向链表的前端
		return elem.Value.(*entry).value // 返回元素的值
	}
	return -1 // 如果元素不在缓存中，返回-1
}

// Put 向缓存中添加一个新元素或更新现有元素
func (this *LRUCache) Put(key int, value int) {

	// 如果元素已经在缓存中
	if elem, found := this.cache[key]; found {
		this.queue.MoveToFront(elem)      // 将元素移动到双向链表的前端
		elem.Value.(*entry).value = value // 更新元素的值
		return
	}

	// 如果缓存已满
	if this.queue.Len() == this.capacity {
		oldest := this.queue.Back()                   // 获取双向链表的最后一个元素
		delete(this.cache, oldest.Value.(*entry).key) // 从哈希表中删除最老的元素
		this.queue.Remove(oldest)                     // 从双向链表中删除最老的元素
	}

	// 在双向链表的前端添加新元素
	// 在哈希表中添加新元素的引用
	elem := this.queue.PushFront(&entry{key, value})
	this.cache[key] = elem

}

// 测试用例
func main() {
	// 示例代码，演示如何使用LRUCache
	lru := Constructor(2)
	lru.Put(1, 1)
	lru.Put(2, 2)
	fmt.Println(lru.Get(1)) // 输出: 1
	lru.Put(3, 3)           // 逐出键2
	fmt.Println(lru.Get(2)) // 输出: -1 (未找到)
	lru.Put(4, 4)           // 逐出键1
	fmt.Println(lru.Get(1)) // 输出: -1 (未找到)
	fmt.Println(lru.Get(3)) // 输出: 3
	fmt.Println(lru.Get(4)) // 输出: 4
}
