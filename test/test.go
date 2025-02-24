package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type NameScore struct {
	Name  string
	Score int
}

func main() {
	dir := "."       // 当前目录，可以根据需要修改
	concurrency := 2 // 并发数

	files, err := getTxtFiles(dir)
	if err != nil {
		fmt.Printf("Error getting .txt files: %v\n", err)
		return
	}

	result := processFiles(files, concurrency)
	top10 := getTop10(result)

	fmt.Println("Top 10 names with highest scores:")
	for _, ns := range top10 {
		fmt.Printf("%s: %d\n", ns.Name, ns.Score)
	}
}

// 获取指定目录下的所有 .txt 文件
func getTxtFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func processFiles(files []string, concurrency int) map[string]int {
	var wg sync.WaitGroup
	fileChan := make(chan string, len(files))
	resultChan := make(chan map[string]int, len(files))
	result := make(map[string]int)
	var mu sync.Mutex

	// 启动goroutine
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				data := readFile(file)
				resultChan <- data
			}
		}()
	}

	// 发送文件到channel
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 合并结果
	for data := range resultChan {
		mu.Lock()
		for name, score := range data {
			result[name] += score
		}
		mu.Unlock()
	}

	return result
}

func readFile(file string) map[string]int {
	data := make(map[string]int)

	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", file, err)
		return data
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // 跳过第一行标题
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]                     // 第一列是name
		score, err := strconv.Atoi(parts[1]) // 第二列是score
		if err != nil {
			fmt.Printf("Error converting score to int in file %s: %v\n", file, err)
			continue
		}
		data[name] += score
	}

	return data
}

func getTop10(data map[string]int) []NameScore {
	var nameScores []NameScore
	for name, score := range data {
		nameScores = append(nameScores, NameScore{Name: name, Score: score})
	}

	sort.Slice(nameScores, func(i, j int) bool {
		return nameScores[i].Score > nameScores[j].Score
	})

	if len(nameScores) > 10 {
		return nameScores[:10]
	}
	return nameScores
}
