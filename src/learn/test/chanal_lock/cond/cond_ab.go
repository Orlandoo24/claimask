package main

import (
	"fmt"
	"sync"
)

func main() {
	// 创建一个WaitGroup，用于等待所有goroutine完成
	var wg sync.WaitGroup

	// 创建一个互斥锁
	var m sync.Mutex

	// 使用互斥锁创建一个条件变量
	c := sync.NewCond(&m)

	// 创建一个布尔变量，用于控制打印的顺序
	toggle := false

	// 增加WaitGroup的计数器，表示有两个goroutine需要等待
	wg.Add(2)

	// 创建第一个goroutine
	go func() {
		// a 循环打印50次
		for i := 1; i <= 50; i++ {
			// 锁定互斥锁
			m.Lock()

			// 等待toggle为true
			for toggle == false {
				// 等待条件变量的信号
				c.Wait()
			}

			// 打印数字
			fmt.Println("线程a打印:", i)

			// 切换toggle的值
			toggle = !toggle

			// 解锁互斥锁
			m.Unlock()

			// 发送条件变量的信号
			c.Signal()
		}

		// goroutine完成后，减少WaitGroup的计数器
		wg.Done()
	}()

	// 创建第二个goroutine
	go func() {
		// b 线程循环打印50次
		for i := 1; i <= 50; i++ {
			// 锁定互斥锁
			m.Lock()

			// 等待toggle为false
			for toggle == true {
				// 等待条件变量的信号
				c.Wait()
			}

			// 打印数字
			fmt.Println("线程b打印:", i)

			// 切换toggle的值
			toggle = !toggle

			// 解锁互斥锁
			m.Unlock()

			// 发送条件变量的信号
			c.Signal()
		}

		// goroutine完成后，减少WaitGroup的计数器
		wg.Done()
	}()

	// 等待所有goroutine完成
	wg.Wait()
}
