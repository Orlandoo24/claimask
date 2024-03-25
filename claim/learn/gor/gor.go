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

// 声明一个互斥锁，用于保护总时间和请求计数
var (
	totalTime time.Duration
	mutex     sync.Mutex
)

func main() {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 记录开始时间
	startTime := time.Now()

	// 并发发起10,000个请求
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			sendRequest()
			wg.Done()
		}()
	}
	wg.Wait()

	// 计算总时间
	totalTime = time.Since(startTime)

	// 输出总时间
	fmt.Printf("所有请求完成\n总时间：%.2f秒\n", totalTime.Seconds())
}

// 发送HTTP请求
func sendRequest() {
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
}

// 生成随机地址
func generateRandomAddress() string {
	// 生成随机数作为地址
	randomNumber := rand.Intn(1000)
	return fmt.Sprintf("Address%d", randomNumber)
}
