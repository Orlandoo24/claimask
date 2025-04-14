package main

//
//import (
//	"bufio"
//	"fmt"
//	"os"
//	"path/filepath"
//	"strings"
//	"sync"
//	"time"
//)
//
//const (
//	outputFile    = "project_code.txt"
//	maxGoroutines = 16 // 最大并发goroutine数
//)
//
//var (
//	excludeDirs  = map[string]bool{".git": true, "__pycache__": true, "node_modules": true}
//	excludeFiles = map[string]bool{".pyc": true, ".DS_Store": true}
//	includeExt   = map[string]bool{".go": true, ".py": true, ".yaml": true, ".md": true}
//)
//
//func main() {
//	startTime := time.Now()
//	defer func() {
//		fmt.Printf("\n扫描完成，耗时: %.2fs\n", time.Since(startTime).Seconds())
//	}()
//
//	// 获取项目根目录
//	projectRoot, _ := os.Getwd()
//	if len(os.Args) > 1 {
//		projectRoot = os.Args[1]
//	}
//
//	// 阶段1：预扫描获取文件列表
//	filePaths := make(chan string, 1000)
//	go scanFiles(projectRoot, filePaths)
//
//	// 初始化进度条
//	total := countFiles(projectRoot)
//	bar := pb.StartNew(total)
//	defer bar.Finish()
//
//	// 阶段2：并发处理
//	var wg sync.WaitGroup
//	outputMutex := &sync.Mutex{}
//	sem := make(chan struct{}, maxGoroutines) // 并发控制
//
//	for path := range filePaths {
//		wg.Add(1)
//		sem <- struct{}{}
//		go func(p string) {
//			defer func() {
//				<-sem
//				wg.Done()
//				bar.Increment()
//			}()
//			processFile(p, projectRoot, outputMutex)
//		}(path)
//	}
//
//	wg.Wait()
//}
//
//// 递归扫描文件并发送到channel
//func scanFiles(root string, paths chan<- string) {
//	defer close(paths)
//
//	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			return nil
//		}
//
//		// 过滤目录
//		if info.IsDir() {
//			if excludeDirs[filepath.Base(path)] {
//				return filepath.SkipDir
//			}
//			return nil
//		}
//
//		// 过滤文件
//		ext := filepath.Ext(path)
//		if excludeFiles[ext] {
//			return nil
//		}
//		if !includeExt[ext] && filepath.Base(path) != "requirements.txt" {
//			return nil
//		}
//
//		paths <- path
//		return nil
//	})
//}
//
//// 处理单个文件
//func processFile(path, root string, mutex *sync.Mutex) {
//	relativePath, _ := filepath.Rel(root, path)
//	header := fmt.Sprintf("\n// ====== FILE: %s ======\n\n", relativePath)
//
//	content, err := readFileWithLineNumbers(path)
//	if err != nil {
//		content = fmt.Sprintf("// Error reading file: %v\n", err)
//	}
//
//	// 加锁写入
//	mutex.Lock()
//	defer mutex.Unlock()
//
//	if err := appendToFile(outputFile, header+content+"\n\n"); err != nil {
//		fmt.Printf("写入失败: %v\n", err)
//	}
//}
//
//// 带行号读取文件
//func readFileWithLineNumbers(path string) (string, error) {
//	file, err := os.Open(path)
//	if err != nil {
//		return "", err
//	}
//	defer file.Close()
//
//	var sb strings.Builder
//	scanner := bufio.NewScanner(file)
//	lineNum := 1
//
//	for scanner.Scan() {
//		sb.WriteString(fmt.Sprintf("%4d | %s\n", lineNum, scanner.Text()))
//		lineNum++
//	}
//
//	if err := scanner.Err(); err != nil {
//		return "", err
//	}
//	return sb.String(), nil
//}
//
//// 原子化追加写入
//func appendToFile(filename, content string) error {
//	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//
//	_, err = f.WriteString(content)
//	return err
//}
//
//// 预扫描统计文件总数
//func countFiles(root string) int {
//	count := 0
//	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
//		if err != nil || info.IsDir() {
//			return nil
//		}
//		count++
//		return nil
//	})
//	return count
//}
