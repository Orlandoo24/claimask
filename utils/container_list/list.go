package main

import (
	"container/list"
	"fmt"
)

func main() {
	// 创建一个新的链表
	l := list.New()

	// 在链表尾部添加元素
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	// 在链表头部添加元素
	l.PushFront(0)

	// 遍历链表并打印元素
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}

	// // 删除链表头部元素
	// l.Remove(l.Front())

	// // 删除链表尾部元素
	// l.Remove(l.Back())

	// // 再次遍历链表并打印元素
	// for e := l.Front(); e != nil; e = e.Next() {
	// 	fmt.Println(e.Value)
	// }
}
