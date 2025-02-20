package main

import (
	"context"
	"fmt"
	"log"
	"time"

	// 分布式锁
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"

	"github.com/Shopify/sarama"                      // 导入 Kafka 客户端库
	"github.com/cloudwego/hertz/pkg/app"             // 导入 Hertz 框架
	"github.com/cloudwego/hertz/pkg/app/server"      // 导入 Hertz 服务器模块
	"github.com/cloudwego/hertz/pkg/protocol/consts" // 导入 Hertz 常量定义

	"github.com/jinzhu/gorm" // 导入 ORM 库
	"github.com/shopspring/decimal"
)

// 收益领取日志表结构体
type RichXRewardLog struct {
	ID           uint64    `gorm:"primary_key;column:id"`
	Address      string    `gorm:"column:address"`
	TotalReward  uint64    `gorm:"column:total_reward"`
	UpdateReward time.Time `gorm:"column:update_reward"`
	Latest       time.Time `gorm:"column:latest"`
	LatestDetail time.Time `gorm:"column:latest_detail"`
	Status       bool      `gorm:"column:status"`
	InsertTime   time.Time `gorm:"column:insert_time"`
	UpdateTime   time.Time `gorm:"column:update_time"`
	IsDeleted    bool      `gorm:"column:is_deleted"`
}

// 收益发放订单表结构体
type RichOrder struct {
	ID         uint64          `gorm:"primary_key;column:id"`
	OrderID    uint64          `gorm:"column:order_id"`
	Address    string          `gorm:"column:address"`
	OrderTime  time.Time       `gorm:"column:order_time"`
	Status     int             `gorm:"column:status"`
	RewardAmt  decimal.Decimal `gorm:"column:reward_amt"`
	InsertTime time.Time       `gorm:"column:insert_time"`
	UpdateTime time.Time       `gorm:"column:update_time"`
	IsDeleted  bool            `gorm:"column:is_deleted"`
}

// 定义请求参数结构体
type (
	RichxParam struct {
		Address string `json:"address"`
	}
)

// 定义全局数据源变量
var (
	DB *gorm.DB // MySQL 数据库连接对象
	// RD *redis.Client

	KafkaProducer sarama.SyncProducer // Kafka 生产者
	KafkaConsumer sarama.Consumer     // Kafka 消费者
)

// 定义全局Kafka配置变量
var kafkaConfig *sarama.Config

// 初始化全局Kafka配置
func init() {
	kafkaConfig = sarama.NewConfig()                            // 创建新的 Kafka 配置对象
	kafkaConfig.Producer.Return.Successes = true                // 设置生产者返回成功的配置
	kafkaConfig.Producer.Retry.Max = 3                          // 设置生产者最大重试次数
	kafkaConfig.Producer.Retry.Backoff = 100 * time.Millisecond // 设置生产者重试间隔
	kafkaConfig.Consumer.Return.Errors = true                   // 设置消费者返回错误的配置
}

// 初始化Kafka生产者
func initKafkaProducer(brokers []string) error {
	producer, err := sarama.NewSyncProducer(brokers, kafkaConfig) // 创建同步生产者
	if err != nil {
		return err
	}
	KafkaProducer = producer // 将生产者赋值给全局变量
	return nil
}

// 初始化Kafka消费者
func initKafkaConsumer(brokers []string, topic string) error {
	consumer, err := sarama.NewConsumer(brokers, kafkaConfig) // 创建消费者
	if err != nil {
		return err
	}
	KafkaConsumer = consumer // 将消费者赋值给全局变量

	partitionConsumer, err := KafkaConsumer.ConsumePartition(topic, 0, sarama.OffsetNewest) // 创建分区消费者
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages(): // 监听分区消费者的消息通道
				fmt.Printf("Received message: %s\n", string(msg.Value))
				// 在这里添加处理消息的逻辑
				// 待定
			case err := <-partitionConsumer.Errors(): // 监听分区消费者的错误通道
				log.Printf("Received consumer error: %s\n", err.Error())
			}
		}
	}()

	return nil
}

func main() {
	var DBerr error
	DB, DBerr = gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True") // 连接 MySQL 数据库
	if DBerr != nil {
		log.Fatalf("failed to connect database: %v", DBerr)
	}

	brokers := []string{"localhost:9092"} // Kafka 集群地址
	topic := "richx"                      // Kafka 主题

	if err := initKafkaProducer(brokers); err != nil { // 初始化 Kafka 生产者
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}

	if err := initKafkaConsumer(brokers, topic); err != nil { // 初始化 Kafka 消费者
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}

	h := server.Default(server.WithHostPorts("127.0.0.1:8890")) // 创建 Hertz 服务器

	// 创建一个redis的客户端连接
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "localhost:6379",
	})
	// 创建redsync的客户端连接池
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// 创建redsync实例
	rs := redsync.New(pool)

	h.POST("/richx", func(c context.Context, ctx *app.RequestContext) {
		param := RichxParam{}
		if bindErr := ctx.Bind(&param); bindErr != nil { // 解析请求参数
			ctx.String(consts.StatusBadRequest, "bind error: %s", bindErr.Error())
			return
		}

		// 1. 查询今日收益领取状态
		var log RichXRewardLog
		if err := DB.Where("address = ? AND DATE(latest) = CURDATE()", param.Address).First(&log).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 用户今日还没有领取收益，可以继续后续步骤
			} else {
				// 数据库查询出错
				ctx.String(consts.StatusInternalServerError, "database error: %s", err.Error())
				return
			}
		} else {
			// 用户今日已经领取了收益，直接返回
			ctx.JSON(consts.StatusOK, map[string]interface{}{
				"message": "You have already claimed your reward today. Please check back tomorrow.",
			})
			return
		}

		// 2. 查询迄今为止累计的收益
		var totalReward uint64
		if err := DB.Model(&RichXRewardLog{}).Where("address = ?", param.Address).Select("total_reward").Row().Scan(&totalReward); err != nil {
			if err == gorm.ErrRecordNotFound {
				// 用户还没有任何收益，可以继续后续步骤
			} else {
				// 数据库查询出错
				ctx.String(consts.StatusInternalServerError, "database error: %s", err.Error())
				return
			}
		} else {
			// 获取到了用户迄今为止的累计收益
			fmt.Printf("User %s has a total reward of %d\n", param.Address, totalReward)
		}

		// 到了数据写操作开始加分布式锁
		// 创建一个互斥锁，设置锁的持有时间为12秒
		mutex := rs.NewMutex("my-global-mutex", redsync.WithExpiry(12*time.Second))
		// 尝试获取锁
		if err := mutex.Lock(); err != nil {
			ctx.String(consts.StatusInternalServerError, "failed to acquire lock: %s", err.Error())
			return
		}

		// 3. 更新收益领取状态更改为正在发放
		log.Status = true // 假设 true 表示 "正在发放"
		if err := DB.Save(&log).Error; err != nil {
			ctx.String(consts.StatusInternalServerError, "database error: %s", err.Error())
			return
		}
		// 4. 将收益重置为0
		log.TotalReward = 0
		if err := DB.Save(&log).Error; err != nil {
			ctx.String(consts.StatusInternalServerError, "database error: %s", err.Error())
			return
		}
		// 5. 更新收益最新领取时间
		log.Latest = time.Now()
		if err := DB.Save(&log).Error; err != nil {
			ctx.String(consts.StatusInternalServerError, "database error: %s", err.Error())
			return
		}
		// 释放锁
		if ok, err := mutex.Unlock(); !ok || err != nil {
			ctx.String(consts.StatusInternalServerError, "failed to release lock: %s", err.Error())
			return
		}

		// 6. 创建收益发放订单
		order := RichOrder{
			Address:   param.Address,
			OrderTime: time.Now(),
			Status:    1, // 假设 1 表示 "正在发放"
			RewardAmt: decimal.NewFromInt(int64(totalReward)),
		}
		if err := DB.Create(&order).Error; err != nil {
			ctx.String(consts.StatusInternalServerError, "database error: %s", err.Error())
			return
		}

		// 7. 将收益发放所需要的数据封装为Kafka消息
		// 构造 message
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(param.Address),
		}

		// 8. 异步发放收益
		// 发送消息到 Kafka
		partition, offset, err := KafkaProducer.SendMessage(msg)
		if err != nil {
			ctx.String(consts.StatusInternalServerError, "failed to send message to Kafka: %s", err.Error())
			return
		}

		// 返回“收益正在分发”的响应
		fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", msg.Topic, partition, offset)
		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"message": "Your reward is being distributed. Please check back later for the status.",
		})
	})

	h.Spin() // 启动服务器
}
