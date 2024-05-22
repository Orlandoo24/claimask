package main

import (
	"fmt"
	"sync"
)

func main() {
	// 创建一个WaitGroup，用于等待所有goroutine完成
	var wg sync.WaitGroup

	// 创建一个带缓冲的通道，缓冲大小为1
	ch := make(chan bool, 1)

	// 增加WaitGroup的计数器，表示有两个goroutine需要等待
	wg.Add(2)

	// 创建第一个goroutine
	go func() {
		// a 循环打印50次
		for i := 1; i <= 50; i++ {
			// 向通道发送一个信号，表示准备打印
			ch <- true

			// 打印数字
			fmt.Println("线程a打印:", i)

			// 从通道接收一个信号，表示打印完成
			<-ch
		}
		// goroutine完成后，减少WaitGroup的计数器
		wg.Done()
	}()

	// 创建第二个goroutine
	go func() {
		// b 线程循环打印50次
		for i := 1; i <= 50; i++ {
			// 向通道发送一个信号，表示准备打印
			ch <- true
			// 打印数字
			fmt.Println("线程b打印:", i)
			// 从通道接收一个信号，表示打印完成
			<-ch
		}
		// goroutine完成后，减少WaitGroup的计数器
		wg.Done()
	}()

	// 等待所有goroutine完成
	wg.Wait()
}
