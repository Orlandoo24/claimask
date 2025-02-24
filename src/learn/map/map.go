package main

import (
	"fmt"
	"time"
)

func main() {

	m := make(map[int]int)

	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 10; j++ {
				m[i] = j
			}
		}(i)
	}

	// 启动 10 个读协程
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 100; j++ {
				a := m[i] // 读操作
				// 打印结果
				fmt.Printf("Goroutine %d, Iteration %d: a = %d\n", i, j, a)

			}
		}(i)
	}

	time.Sleep(2 * time.Second) // 等待协程执行完毕

}
