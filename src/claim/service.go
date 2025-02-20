package claim

import (
	"errors"
	"fmt"
	"log"
	"time"

	"claimask/src/claim/model"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
)

// OrderService 订单服务接口
type OrderService interface {
	ClaimPrize() error
	CreateOrder(address string) error
	QueryPrizes() (int, error)
	InitPrizes(quantity int)
}

// OrderServiceImpl 订单服务实现
type OrderServiceImpl struct {
	OrderDAO OrderDAO
	RD       *redis.Client
}

// NewOrderService 创建新的订单服务实例
func NewOrderService(orderDAO OrderDAO, rd *redis.Client) OrderService {
	return &OrderServiceImpl{
		OrderDAO: orderDAO,
		RD:       rd,
	}
}

// ClaimPrize 用于领取奖品的函数，传入一个 Redis 客户端 RD，返回可能的错误
func (s *OrderServiceImpl) ClaimPrize() error {
	var prizes int                 // 声明奖品数量变量
	maxRetries := 3                // 最大重试次数
	maxDuration := 5 * time.Second // 最大执行时间限制，假设为5秒

	retries := 0        // 初始化重试次数为0
	start := time.Now() // 记录开始时间

	for {
		err := s.RD.Watch(func(tx *redis.Tx) error {
			var err error

			// 从 Redis 中获取奖品数量
			prizes, err = tx.Get("prizes").Int()
			if err != nil {
				return err
			}

			// 如果奖品数量大于 0，则递减奖品数量
			if prizes > 0 {
				_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
					// 在 Redis 中递减奖品数量
					pipe.Decr("prizes")
					return nil
				})
				if err != nil {
					return err
				}
				return nil
			}
			return errors.New("奖品已经领完了") // 如果奖品数量为0，则返回错误信息
		}, "prizes")

		if err == nil {
			// 如果没有错误，表示奖品数量获取和递减成功，跳出循环
			break
		} else if err == redis.TxFailedErr {
			fmt.Print("当前有其他事务对 prizes 键进行了修改，事务回滚，并进行重试")

			// 如果出现 redis.TxFailedErr 错误，表示事务失败，需要重试
			retries++
			if retries >= maxRetries || time.Since(start) >= maxDuration {
				// 如果达到最大重试次数或者超过最大时间限制，退出循环并返回错误信息
				return errors.New("重试次数超过限制或执行时间超时")
			}
			continue // Retry
		} else {
			// 如果出现其他错误，返回给客户端错误信息，并结束处理
			return err
		}
	}
	return nil
}

// CreateOrder 创建订单
func (s *OrderServiceImpl) CreateOrder(address string) error {
	order := &model.Order{
		OrderID:    generateOrderID(),
		Address:    address,
		Json:       `{"key": "value"}`,
		InsertTime: time.Now(),
		UpdateTime: time.Now(),
	}

	return s.OrderDAO.CreateOrder(order)
}

// generateOrderID 生成分布式ID
func generateOrderID() uint64 {
	// 创建一个新的节点（Node），用于生成雪花ID
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("无法创建雪花节点: %v", err)
	}

	// 生成一个新的雪花ID
	id := node.Generate()

	// 将 int64 类型的 ID 转换为 uint64 类型
	return uint64(id.Int64())
}

// QueryPrizes 查询奖品数量
func (s *OrderServiceImpl) QueryPrizes() (int, error) {
	return s.RD.Get("prizes").Int()
}

// InitPrizes 初始化奖品数量
func (s *OrderServiceImpl) InitPrizes(quantity int) {
	s.RD.Set("prizes", quantity, 0)
}
