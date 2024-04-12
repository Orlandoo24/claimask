package main

import (
	"fmt"
)

func main() {
	for i := 1; i <= 100; i++ {
		go fmt.Println(i)
	}
}
