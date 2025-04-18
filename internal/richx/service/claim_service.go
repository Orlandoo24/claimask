package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"

	"claimask/internal/richx/model"
)

const redisLockKey = "richx_test:claim:lock:%s"

// DB是数据库服务
type DB struct {
	*gorm.DB
}

// RedisClient是Redis客户端
type RedisClient struct {
	client *redis.Client
}

// KafkaProducer是Kafka生产者
type KafkaProducer struct {
	producer sarama.SyncProducer
	enabled  bool
}

// ClaimService 定义领取服务接口
type ClaimService struct {
	db            *DB
	redisClient   *RedisClient
	kafkaProducer *KafkaProducer
	logger        *log.Logger
}

// NewClaimService 创建领取服务实例
func NewClaimService(db *DB, redisClient *RedisClient, kafkaProducer *KafkaProducer, logger *log.Logger) *ClaimService {
	return &ClaimService{
		db:            db,
		redisClient:   redisClient,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

// InitDB 初始化数据库连接
func InitDB() (*DB, error) {
	db, err := gorm.Open("mysql", "root:123@jiaru@tcp(localhost:3306)/richx_test?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	return &DB{db}, nil
}

// InitRedis 初始化Redis连接
func InitRedis() (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &RedisClient{client: client}, nil
}

// InitKafka 初始化Kafka连接
func InitKafka() (*KafkaProducer, error) {
	// 由于Kafka未安装，返回一个Mock的Kafka生产者
	return &KafkaProducer{
		producer: nil,
		enabled:  false,
	}, nil
}

// AcquireLock 获取分布式锁
func (r *RedisClient) AcquireLock(ctx context.Context, address string, expiration time.Duration) (bool, error) {
	lockKey := fmt.Sprintf(redisLockKey, address)
	return r.client.SetNX(lockKey, "1", expiration).Result()
}

// ReleaseLock 释放分布式锁
func (r *RedisClient) ReleaseLock(ctx context.Context, address string) error {
	lockKey := fmt.Sprintf(redisLockKey, address)
	return r.client.Del(lockKey).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// CanClaim 判断用户是否可以领取收益
func (s *ClaimService) CanClaim(address string) (bool, *big.Float, error) {
	var rewardLog model.RichRewardLog
	if err := s.db.Where("address = ?", address).First(&rewardLog).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, big.NewFloat(0), nil
		}
		return false, nil, err
	}

	// 如果总收益为0，则无法领取
	zero := big.NewFloat(0)
	if rewardLog.TotalReward.Cmp(zero) == 0 {
		return false, zero, nil
	}

	today := time.Now().Truncate(24 * time.Hour)

	// 如果从未领取过，可以领取
	if rewardLog.Latest == nil {
		return true, rewardLog.TotalReward, nil
	}

	// 如果最近一次领取是今天，不能领取
	latestDate := rewardLog.Latest.Truncate(24 * time.Hour)
	if latestDate.Equal(today) {
		return false, zero, nil
	}

	// 可以领取
	return true, rewardLog.TotalReward, nil
}

// CreateOrder 创建领取订单
func (s *ClaimService) CreateOrder(address string, amount *big.Float) (int64, error) {
	orderID := time.Now().UnixNano()
	order := model.ClaimOrder{
		OrderID:   orderID,
		Address:   address,
		Amount:    amount,
		Status:    1, // 创建状态
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(&order).Error; err != nil {
		return 0, err
	}

	return orderID, nil
}

// SendToKafka 发送消息到Kafka
func (s *KafkaProducer) SendToKafka(message model.KafkaMessage) error {
	if !s.enabled {
		return nil // Mock实现，直接返回成功
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "rich-claim",
		Value: sarama.StringEncoder(msgBytes),
		Key:   sarama.StringEncoder(message.Address), // 使用地址作为Key确保同一用户的消息顺序处理
	}

	_, _, err = s.producer.SendMessage(msg)
	return err
}

// CalculateSha256 计算SHA256哈希
func CalculateSha256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Claim 处理用户领取请求
func (s *ClaimService) Claim(ctx context.Context, address string) (*model.Response, error) {
	s.logger.Printf("收到来自 %s 的领取请求", address)

	// 使用Redis分布式锁防止并发请求
	acquired, err := s.redisClient.AcquireLock(ctx, address, 10*time.Second)
	if err != nil {
		s.logger.Printf("获取锁失败: %v", err)
		return &model.Response{
			Code:    500,
			Message: "系统错误，请稍后再试",
		}, err
	}

	if !acquired {
		s.logger.Printf("用户 %s 操作太频繁", address)
		return &model.Response{
			Code:    429,
			Message: "操作太频繁，请稍后再试",
		}, nil
	}

	defer func() {
		if err := s.redisClient.ReleaseLock(ctx, address); err != nil {
			s.logger.Printf("释放锁失败: %v", err)
		}
	}()

	// 检查是否可以领取
	canClaim, amount, err := s.CanClaim(address)
	if err != nil {
		s.logger.Printf("检查是否可领取失败: %v", err)
		return &model.Response{
			Code:    500,
			Message: "系统错误，请稍后再试",
		}, err
	}

	if !canClaim {
		s.logger.Printf("用户 %s 当天已领取或无可领取收益", address)
		return &model.Response{
			Code:    400,
			Message: "您今天已领取或没有可领取的收益",
		}, nil
	}

	// 创建订单
	orderID, err := s.CreateOrder(address, amount)
	if err != nil {
		s.logger.Printf("创建订单失败: %v", err)
		return &model.Response{
			Code:    500,
			Message: "系统错误，请稍后再试",
		}, err
	}

	// 更新领取记录
	now := time.Now()
	if err := s.db.Model(&model.RichRewardLog{}).
		Where("address = ?", address).
		Update("latest", now).Error; err != nil {
		s.logger.Printf("更新领取记录失败: %v", err)
		// 继续处理，不影响用户体验
	}

	// 构造Kafka消息
	message := model.KafkaMessage{
		OrderID:   strconv.FormatInt(orderID, 10),
		Address:   address,
		Amount:    amount,
		Timestamp: time.Now(),
		Sha:       CalculateSha256(address + amount.String() + strconv.FormatInt(time.Now().Unix(), 10)),
	}

	// 发送到Kafka
	if err := s.kafkaProducer.SendToKafka(message); err != nil {
		s.logger.Printf("发送到Kafka失败: %v", err)
		// 继续处理，假设有重试机制
	}

	// 异步处理（模拟）
	go s.ProcessClaimAsync(message)

	return &model.Response{
		Code: 200,
		Data: map[string]interface{}{
			"orderID": orderID,
			"amount":  amount.String(),
		},
		Message: "领取成功",
	}, nil
}

// ProcessClaimAsync 异步处理领取请求
func (s *ClaimService) ProcessClaimAsync(message model.KafkaMessage) error {
	s.logger.Printf("异步处理订单 %s", message.OrderID)

	// 模拟处理延迟
	time.Sleep(2 * time.Second)

	// 解析订单ID
	orderID, err := strconv.ParseInt(message.OrderID, 10, 64)
	if err != nil {
		s.logger.Printf("解析订单ID失败: %v", err)
		return err
	}

	// 更新订单状态为处理中
	if err := s.db.Model(&model.ClaimOrder{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     2,
			"updated_at": time.Now(),
		}).Error; err != nil {
		s.logger.Printf("更新订单状态失败: %v", err)
		return err
	}

	// 模拟链上处理
	time.Sleep(3 * time.Second)

	// 更新订单状态为已确认
	if err := s.db.Model(&model.ClaimOrder{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     3,
			"updated_at": time.Now(),
		}).Error; err != nil {
		s.logger.Printf("更新订单状态失败: %v", err)
		return err
	}

	s.logger.Printf("订单 %s 处理完成", message.OrderID)
	return nil
}
