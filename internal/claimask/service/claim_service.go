package service

import (
	"astro-orderx/internal/claimask/dao"
	"astro-orderx/internal/claimask/model/po"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
)

// 定义包级别常量
const (
	maxClaimRetries  = 3
	maxClaimDuration = 5 * time.Second
	redisKeyPrizes   = "prizes"
	defaultNodeID    = 1
)

var (
	// 定义明确错误类型方便上层处理
	ErrNoPrizeLeft       = errors.New("no prize left")
	ErrExceedMaxAttempts = errors.New("exceed max attempts")
)

// ClaimService 定义订单服务接口
type ClaimService interface {
	ClaimPrize() error
	ClaimPrizeV2() error
	CreateOrder(address string) error
	QueryPrizes() (int, error)
	InitPrizes(quantity int)
}

// ClaimServiceImpl 实现订单服务接口
type ClaimServiceImpl struct {
	orderDAO dao.OrderDAO
	redisCli *redis.Client
}

// NewClaimService 创建订单服务实例
func NewClaimService(orderDAO dao.OrderDAO, rd *redis.Client) ClaimService {
	return &ClaimServiceImpl{
		orderDAO: orderDAO,
		redisCli: rd,
	}
}

// ClaimPrize 领取奖品（带事务重试机制）
// 实现原理：
// 1. 使用Redis Watch实现乐观锁控制
// 2. 采用有限次数的重试机制处理事务冲突
// 3. 设置总操作超时时间防止长时间阻塞
// 返回值：
//   - 成功时返回nil
//   - ErrNoPrizeLeft 奖品已领完
//   - ErrExceedMaxAttempts 超过最大尝试次数
func (s *ClaimServiceImpl) ClaimPrize() error {
	startTime := time.Now()

	// 有限重试循环（避免无限重试导致系统阻塞）
	for retry := 0; retry < maxClaimRetries; retry++ {
		// 超时检查：总操作时间超过最大允许时长则立即终止
		if time.Since(startTime) > maxClaimDuration {
			return fmt.Errorf("operation timeout: %w", ErrExceedMaxAttempts)
		}

		// 开启Redis事务监控
		err := s.redisCli.Watch(func(tx *redis.Tx) error {
			// --- 事务开始 ---
			// 原子化操作步骤：
			// 1. 获取当前奖品数量
			// 2. 检查库存有效性
			// 3. 执行库存递减

			// 步骤1：获取当前奖品数量
			prizeCount, err := tx.Get(redisKeyPrizes).Int()
			if err != nil && err != redis.Nil { // 处理非"key不存在"的其他错误
				return fmt.Errorf("get prize count failed: %w", err)
			}

			// 步骤2：库存检查
			// 当库存<=0时返回特定错误，终止事务流程
			if prizeCount <= 0 {
				return ErrNoPrizeLeft
			}

			// 步骤3：执行库存递减操作
			// 使用管道提升事务执行效率（单次网络往返）
			_, err = tx.TxPipelined(func(pipe redis.Pipeliner) error {
				pipe.Decr(redisKeyPrizes)
				return nil
			})
			return err
		}, redisKeyPrizes) // 监控prizes键的变化

		// --- 事务处理结果分析 ---
		switch {
		case err == nil:
			// 成功情况：事务执行成功，直接返回
			return nil
		case errors.Is(err, ErrNoPrizeLeft):
			// 业务终止情况：明确无库存，直接向上返回错误
			return err
		case errors.Is(err, redis.TxFailedErr):
			// 事务冲突情况：记录日志并继续重试
			log.Printf("transaction conflict detected, retry count: %d/%d",
				retry+1, maxClaimRetries)
			continue
		default:
			// 不可恢复错误：包装错误信息后返回
			return fmt.Errorf("unexpected error: %w", err)
		}
	}

	// 重试耗尽：返回明确的尝试次数超限错误
	return ErrExceedMaxAttempts
}

// ClaimPrizeV2 领取奖品（利用 Redis 的单线程特性保证数据一致性）
// 实现原理：
// 1. 使用 Redis 的原子操作 DECR 保证库存递减的原子性。
// 2. 在 DECR 之前检查库存，避免超卖。
// 返回值：
//   - 成功时返回 nil
//   - ErrNoPrizeLeft 奖品已领完
func (s *ClaimServiceImpl) ClaimPrizeV2() error {
	// 获取当前库存
	prizeCount, err := s.redisCli.Get(redisKeyPrizes).Int()
	if err != nil && err != redis.Nil { // 处理非"key不存在"的其他错误
		return fmt.Errorf("get prize count failed: %w", err)
	}

	// 检查库存
	if prizeCount <= 0 {
		return ErrNoPrizeLeft
	}

	// 执行库存递减操作（原子操作）
	newCount, err := s.redisCli.Decr(redisKeyPrizes).Result()
	if err != nil {
		return fmt.Errorf("decr prize count failed: %w", err)
	}
	// 再次检查库存，避免超卖
	if newCount < 0 {
		// 如果库存减为负数，回滚操作
		_, _ = s.redisCli.Incr(redisKeyPrizes).Result()
		return ErrNoPrizeLeft
	}
	return nil
}

// CreateOrder 创建订单
func (s *ClaimServiceImpl) CreateOrder(address string) error {
	orderID, err := generateOrderID()
	if err != nil {
		return fmt.Errorf("generate order id failed: %w", err)
	}

	order := &po.Order{
		OrderID:    orderID,
		Address:    address,
		Json:       `{"key": "value"}`, // 建议改为配置项或参数传入
		InsertTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := s.orderDAO.CreateOrder(order); err != nil {
		return fmt.Errorf("create order failed: %w", err)
	}
	return nil
}

// generateOrderID 生成分布式唯一ID
func generateOrderID() (uint64, error) {
	node, err := snowflake.NewNode(defaultNodeID)
	if err != nil {
		return 0, fmt.Errorf("create snowflake node failed: %w", err)
	}
	return uint64(node.Generate().Int64()), nil
}

// QueryPrizes 查询当前奖品数量
func (s *ClaimServiceImpl) QueryPrizes() (int, error) {
	count, err := s.redisCli.Get(redisKeyPrizes).Int()
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("query prizes failed: %w", err)
	}
	return count, nil
}

// InitPrizes 初始化奖品数量
func (s *ClaimServiceImpl) InitPrizes(quantity int) {
	if _, err := s.redisCli.Set(redisKeyPrizes, quantity, 0).Result(); err != nil {
		log.Printf("init prizes failed: %v", err)
	}
}
