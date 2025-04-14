package model

import (
	"math/big"
	"time"
)

// RichRewardLog 对应Java中的RichRewardLogDo
type RichRewardLog struct {
	ID           uint       `gorm:"primary_key"`
	Address      string     `gorm:"column:address;type:varchar(64)"`
	TotalReward  *big.Float `gorm:"column:total_reward;type:decimal(65,18)"`
	UpdateReward time.Time  `gorm:"column:update_reward"`
	Latest       *time.Time `gorm:"column:latest"` // 最后领取时间
}

// TableName 设置RichRewardLog表名
func (RichRewardLog) TableName() string {
	return "rich_reward_log"
}

// ClaimOrder 对应Java中的OrderDo
type ClaimOrder struct {
	ID        uint       `gorm:"primary_key"`
	OrderID   int64      `gorm:"column:order_id;unique"`
	Address   string     `gorm:"column:address;type:varchar(64)"`
	Amount    *big.Float `gorm:"column:amount;type:decimal(65,18)"`
	Status    int        `gorm:"column:status"` // 1:创建 2:处理中 3:已确认
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

// TableName 设置ClaimOrder表名
func (ClaimOrder) TableName() string {
	return "order"
}

// ClaimRequest 领取请求
type ClaimRequest struct {
	Address string `json:"address" binding:"required"`
}

// KafkaMessage Kafka消息结构
type KafkaMessage struct {
	OrderID   string     `json:"orderId"`
	Address   string     `json:"address"`
	Amount    *big.Float `json:"amount"`
	Timestamp time.Time  `json:"timestamp"`
	Sha       string     `json:"sha"`
}

// ChainClaimRequest 链上请求结构
type ChainClaimRequest struct {
	Address string     `json:"address"`
	Amount  *big.Float `json:"amount"`
	OrderID string     `json:"orderId"`
	Sha     string     `json:"sha"`
}

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	TraceID string      `json:"traceId,omitempty"`
}
