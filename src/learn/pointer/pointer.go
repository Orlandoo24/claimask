package main

import "fmt"

// 定义一个大型数据结构
type LargeStruct struct {
	Data [1000000]int // 一个包含100万个整数的数组
}

// 函数接收 LargeStruct 结构体的值
func processStruct(s LargeStruct) {
	// 对结构体进行一些处理
	// 这里的 s 是函数参数，传递的是结构体的值的副本
	// 任何对 s 的修改都不会影响原始数据
	// 例如，如果我们尝试修改 s 的字段值，只会影响函数内的副本，而不会影响原始数据
	// s.Data[0] = 5 // 这里的修改不会影响原始数据
	fmt.Println("Processing struct with value:", s)
}

// 函数接收 LargeStruct 结构体的指针
func processStructPtr(s *LargeStruct) {
	// 对结构体进行一些处理
	// 这里的 s 是结构体的指针，可以直接操作原始数据
	// 任何对 s 的修改都会影响原始数据
	// 例如，如果我们尝试修改 s 的字段值，会直接影响原始数据
	// s.Data[0] = 5 // 这里的修改会影响原始数据
	fmt.Println("Processing struct with pointer:", *s)
}

func main() {
	// 创建一个大型数据结构的实例
	largeData := LargeStruct{}

	// 使用值传递调用函数
	processStruct(largeData)
	// 在这个调用中，processStruct 函数接收的是 largeData 的副本，所以在函数内部对 largeData 的修改不会影响原始数据

	// 使用指针传递调用函数
	processStructPtr(&largeData)
	// 在这个调用中，processStructPtr 函数接收的是 largeData 的指针，因此可以直接修改原始数据

	// 通过指针修改原始数据
	largeData.Data[0] = 10
	// 这里直接修改了 largeData 的第一个元素的值

	fmt.Println("Value of first element after modification:", largeData.Data[0])
}
