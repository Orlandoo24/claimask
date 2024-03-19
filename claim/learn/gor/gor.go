package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// 请求的目标地址
const targetURL = "http://127.0.0.1:8870/claim"

// 声明一个互斥锁，用于保护请求计数器和总时间
var (
	counterMutex sync.Mutex
	requestCount int
	totalTime    time.Duration
)

func main() {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 并发发起10,000个请求
	var wg sync.WaitGroup
	for i := 0; i < 20000; i++ {
		wg.Add(1)
		go func() {
			sendRequest()
			wg.Done()
		}()
	}
	wg.Wait()

	// 输出总时间
	fmt.Printf("所有请求完成\n总时间：%.2f秒\n", totalTime.Seconds())
}

// 发送HTTP请求
func sendRequest() {
	startTime := time.Now()

	// 生成随机地址
	randomAddress := generateRandomAddress()

	// 创建请求参数
	param := map[string]string{"address": randomAddress}
	paramBytes, err := json.Marshal(param)
	if err != nil {
		fmt.Printf("生成请求参数失败：%v\n", err)
		return
	}

	// 发送POST请求到目标地址
	resp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(paramBytes))
	if err != nil {
		fmt.Printf("发送请求失败：%v\n", err)
		return
	}
	defer resp.Body.Close()

	// 记录请求完成时间并增加请求计数器
	endTime := time.Now()
	increaseRequestCount()
	addTotalTime(endTime.Sub(startTime))

	// 打印请求结果
	fmt.Printf("请求地址：%s，状态码：%d\n", randomAddress, resp.StatusCode)
}

// 生成随机地址
func generateRandomAddress() string {
	// 生成随机数作为地址
	randomNumber := rand.Intn(1000)
	return fmt.Sprintf("Address%d", randomNumber)
}

// 增加请求计数器
func increaseRequestCount() {
	// 使用互斥锁保护请求计数器
	counterMutex.Lock()
	defer counterMutex.Unlock()
	requestCount++
}

// 增加总时间
func addTotalTime(duration time.Duration) {
	// 使用互斥锁保护总时间
	counterMutex.Lock()
	defer counterMutex.Unlock()
	totalTime += duration
}
