package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	// 使用 gorm.Open 函数打开一个数据库连接
	// "mysql" 是驱动名称，"user:password@/dbname?charset=utf8&parseTime=True&loc=Local" 是数据源名称，需要替换为实际的用户名、密码和数据库名
	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic(err) // 如果打开数据库连接出错，直接 panic
	}

	defer db.Close() // 在 main 函数结束时关闭数据库连接

	// 创建一个新的 Ticker，每隔 5 秒就会向其通道发送一次时间
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // 在 main 函数结束时停止 Ticker

	for {
		select {
		case <-ticker.C: // 当 Ticker 的通道接收到时间时
			// 执行 SQL 语句，"YOUR_SQL_STATEMENT" 需要替换为实际的 SQL 语句
			err := db.Exec("update `good` set `num` = `num` + 1 where `order_id` = 123456789").Error
			if err != nil {
				// 如果执行 SQL 语句出错，打印错误信息
				fmt.Println("Error executing SQL statement:", err)
			} else {
				// 如果执行 SQL 语句成功，打印成功信息
				fmt.Println("SQL statement executed successfully")
			}
		}
	}
}
