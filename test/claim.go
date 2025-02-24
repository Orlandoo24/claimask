package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL     = "http://127.0.0.1:8870/claim" // 修复了 URL 的格式问题
	numRequests = 100000
	numWorkers  = 24
)

func generateRandomAddress() string {
	// 生成一个随机的 20 字节数据，模拟以太坊地址
	randomBytes := make([]byte, 20)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return "0x" + hex.EncodeToString(randomBytes)
}

func sendRequest(address string) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", baseURL, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Host", "127.0.0.1:8870")
	req.Header.Set("Connection", "keep-alive")

	// 设置请求体
	body := fmt.Sprintf(`{"address": "%s"}`, address)
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(body)))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Request with address %s completed with status %d\n", address, resp.StatusCode)
}

func worker(wg *sync.WaitGroup, jobs <-chan string) {
	defer wg.Done()
	for address := range jobs {
		sendRequest(address)
	}
}

func main() {
	var wg sync.WaitGroup
	addresses := make(map[string]struct{})
	jobs := make(chan string, numRequests)

	// 生成不重复的 address
	for len(addresses) < numRequests {
		address := generateRandomAddress()
		addresses[address] = struct{}{}
	}

	// 启动 worker
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, jobs)
	}

	// 将任务放入队列
	startTime := time.Now()
	for address := range addresses {
		jobs <- address
	}
	close(jobs)

	// 等待所有 worker 完成
	wg.Wait()
	duration := time.Since(startTime)
	fmt.Printf("Sent %d requests in %.3f seconds\n", numRequests, duration.Seconds())
}
