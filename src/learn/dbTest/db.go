package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Order struct {
	ID         uint      `gorm:"primary_key"`
	OrderID    uint64    `gorm:"column:order_id"`
	Address    string    `gorm:"column:address"`
	Json1      string    `gorm:"column:json1"`
	InsertTime time.Time `gorm:"column:insert_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func main() {
	// 连接到 MySQL 数据库
	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True")
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	defer db.Close()

	// 指定表名为 "order_id"
	db.Table("order_id").AutoMigrate(&Order{})

	// 创建订单
	order := &Order{
		OrderID:    123456789,
		Address:    "Test Address",
		Json1:      `{"key": "value"}`,
		InsertTime: time.Now(), // 分配当前时间戳值
		UpdateTime: time.Now(), // 分配当前时间戳值
	}

	// 插入订单
	dbErr := db.Table("order_id").Create(order).Error
	if dbErr != nil {
		log.Fatalf("无法插入订单: %v", err)
	}

	fmt.Println("数据插入成功!")
}
