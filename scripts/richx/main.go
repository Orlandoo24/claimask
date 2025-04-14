package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// 配置常量
const (
	// Redis相关配置
	redisLockExpiration = 30 * time.Second
	redisLockKey        = "claim:lock:%s" // 用户地址作为锁的后缀

	// Kafka相关配置
	kafkaTopic       = "claim-rewards-topic"
	kafkaMaxRetry    = 3
	kafkaBrokerAddrs = "localhost:9092"

	// 链上交易API
	chainClaimURL = "http://94.130.49.158:3000/claim"
	saltValue     = "CFIrxG7nDq4h2TofxTGlmm220E7UI2JBxf"

	// 分布式锁尝试获取的超时时间
	lockTimeout = 5 * time.Second
)

// 定义模型结构

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

// 请求响应结构

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	TraceID string      `json:"traceId,omitempty"`
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

// 依赖服务接口

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

// 初始化函数
func initDB() (*DB, error) {
	db, err := gorm.Open("mysql", "root:123@jiaru@tcp(localhost:3306)/richx?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	return &DB{db}, nil
}

func initRedis() (*RedisClient, error) {
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

func initKafka() (*KafkaProducer, error) {
	// 由于Kafka未安装，返回一个Mock的Kafka生产者
	return &KafkaProducer{
		producer: nil,
		enabled:  false,
	}, nil
}

// 分布式锁实现
func (r *RedisClient) AcquireLock(ctx context.Context, address string, expiration time.Duration) (bool, error) {
	lockKey := fmt.Sprintf(redisLockKey, address)
	return r.client.SetNX(lockKey, "1", expiration).Result()
}

func (r *RedisClient) ReleaseLock(ctx context.Context, address string) error {
	lockKey := fmt.Sprintf(redisLockKey, address)
	return r.client.Del(lockKey).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// 服务方法实现

// CanClaim 判断用户是否可以领取收益
func (s *ClaimService) CanClaim(address string) (bool, *big.Float, error) {
	var rewardLog RichRewardLog
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
	order := ClaimOrder{
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

// SendToKafka 发送领取消息到Kafka
func (s *KafkaProducer) SendToKafka(message KafkaMessage) error {
	if !s.enabled {
		// Kafka未启用，返回空错误
		return nil
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 创建Kafka消息
	msg := &sarama.ProducerMessage{
		Topic: kafkaTopic,
		Value: sarama.StringEncoder(msgBytes),
		Key:   sarama.StringEncoder(message.Address), // 使用地址作为Key确保同一用户的消息顺序处理
	}

	// 发送消息
	_, _, err = s.producer.SendMessage(msg)
	return err
}

// CalculateSha256 计算SHA256哈希
func CalculateSha256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Claim 处理领取请求
func (s *ClaimService) Claim(ctx context.Context, address string) (*Response, error) {
	// 1. 尝试获取分布式锁
	ctxWithTimeout, cancel := context.WithTimeout(ctx, lockTimeout)
	defer cancel()

	acquired, err := s.redisClient.AcquireLock(ctxWithTimeout, address, redisLockExpiration)
	if err != nil {
		return &Response{Code: 500, Message: "获取锁失败: " + err.Error()}, err
	}

	if !acquired {
		return &Response{Code: 429, Message: "您有一个正在处理的领取请求，请稍后再试"}, nil
	}

	defer func() {
		if err := s.redisClient.ReleaseLock(ctx, address); err != nil {
			s.logger.Printf("释放锁失败: %v", err)
		}
	}()

	// 2. 判断用户是否有领取资格
	canClaim, reward, err := s.CanClaim(address)
	if err != nil {
		return &Response{Code: 500, Message: "查询领取资格失败: " + err.Error()}, err
	}

	if !canClaim {
		return &Response{Code: 403, Message: "您今天已经领取过收益或没有可领取的收益"}, nil
	}

	// 3. 创建订单
	orderID, err := s.CreateOrder(address, reward)
	if err != nil {
		return &Response{Code: 500, Message: "创建订单失败: " + err.Error()}, err
	}

	// 4. 准备Kafka消息
	orderIDStr := strconv.FormatInt(orderID, 10)
	message := KafkaMessage{
		OrderID:   orderIDStr,
		Address:   address,
		Amount:    reward,
		Timestamp: time.Now(),
	}

	// 添加SHA256签名
	claimData := fmt.Sprintf(`{"address":"%s","amount":%s,"orderId":"%s"}%s`,
		address, reward.Text('f', 18), orderIDStr, saltValue)
	message.Sha = CalculateSha256(claimData)

	// 5. 发送到Kafka或直接处理
	if s.kafkaProducer.enabled {
		if err := s.kafkaProducer.SendToKafka(message); err != nil {
			return &Response{Code: 500, Message: "将领取请求加入队列失败: " + err.Error()}, err
		}
	} else {
		// Kafka未启用，直接处理
		s.logger.Printf("Kafka未启用，直接处理领取请求")
		go func() {
			if err := s.ProcessClaimAsync(message); err != nil {
				s.logger.Printf("处理领取请求失败: %v", err)
			}
		}()
	}

	// 6. 返回成功响应
	return &Response{
		Code:    200,
		Message: "领取请求已接收，正在处理中",
		Data: map[string]interface{}{
			"orderID": orderIDStr,
		},
	}, nil
}

// ProcessClaimAsync 异步处理领取请求 (Kafka消费者会调用)
func (s *ClaimService) ProcessClaimAsync(message KafkaMessage) error {
	// 链上请求
	chainRequest := ChainClaimRequest{
		Address: message.Address,
		Amount:  message.Amount,
		OrderID: message.OrderID,
		Sha:     message.Sha,
	}

	// 假设向链上API发送请求
	s.logger.Printf("模拟向链上发送领取请求: %+v", chainRequest)

	// 假设请求成功，更新订单状态
	orderID, _ := strconv.ParseInt(message.OrderID, 10, 64)
	if err := s.db.Model(&ClaimOrder{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     3, // 已确认状态
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	// 更新用户领取记录
	now := time.Now()
	if err := s.db.Model(&RichRewardLog{}).
		Where("address = ?", message.Address).
		Updates(map[string]interface{}{
			"total_reward": big.NewFloat(0),
			"latest":       now,
		}).Error; err != nil {
		return err
	}

	s.logger.Printf("用户 %s 的收益处理完成", message.Address)
	return nil
}

// 控制器实现
func setupRouter(claimService *ClaimService) *gin.Engine {
	r := gin.Default()

	r.POST("/rich/claim", func(c *gin.Context) {
		var req ClaimRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Code:    400,
				Message: "请求参数错误: " + err.Error(),
			})
			return
		}

		resp, err := claimService.Claim(c, req.Address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, resp)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	r.GET("/rich/claim", func(c *gin.Context) {
		address := c.Query("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, Response{
				Code:    400,
				Message: "地址参数不能为空",
			})
			return
		}

		resp, err := claimService.Claim(c, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, resp)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	return r
}

func main() {
	// 初始化日志
	logger := log.New(log.Writer(), "ClaimService: ", log.LstdFlags)

	// 初始化数据库连接
	db, err := initDB()
	if err != nil {
		logger.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	// 初始化Redis客户端
	redisClient, err := initRedis()
	if err != nil {
		logger.Fatalf("初始化Redis失败: %v", err)
	}
	defer redisClient.Close()

	// 初始化Kafka生产者（Mock）
	kafkaProducer, err := initKafka()
	if err != nil {
		logger.Fatalf("初始化Kafka失败: %v", err)
	}

	// 创建服务
	claimService := NewClaimService(db, redisClient, kafkaProducer, logger)

	// 设置路由
	r := setupRouter(claimService)

	// 启动HTTP服务器
	logger.Printf("服务启动在 http://localhost:8881")
	if err := r.Run(":8881"); err != nil {
		logger.Fatalf("启动服务失败: %v", err)
	}
}
