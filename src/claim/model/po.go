package model

import (
	"time"
)

// ClaimParam defines the structure for claim request parameters
type ClaimParam struct {
	Address string `json:"address"`
}

// Order 定义订单结构体
type Order struct {
	ID         uint      `gorm:"primary_key"`
	OrderID    uint64    `gorm:"column:order_id"`
	Address    string    `gorm:"column:address"`
	Json       string    `gorm:"column:json"`
	InsertTime time.Time `gorm:"column:insert_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}
