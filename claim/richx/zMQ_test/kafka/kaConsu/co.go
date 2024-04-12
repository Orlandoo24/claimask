package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/jinzhu/gorm"
)

// 定义全局数据源变量
var (
	DB            *gorm.DB            // MySQL 数据库连接对象
	KafkaProducer sarama.SyncProducer // Kafka 生产者
	KafkaConsumer sarama.Consumer     // Kafka 消费者
)

type KafkaTest struct {
	gorm.Model
	Address string `gorm:"type:varchar(64);not null;comment:用户钱包地址"`
	Status  int    `gorm:"type:bit(1);not null;default:0;comment:今日是否已经领取，0未领取，1已领取"`
}

// 定义请求参数结构体
type (
	RichxParam struct {
		Address string `json:"address"`
		// 可以添加更多VIP用户的参数
	}
)

// 定义全局Kafka配置变量
var kafkaConfig *sarama.Config

// 初始化全局Kafka配置
func init() {
	kafkaConfig = sarama.NewConfig()

	// 生产者配置
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Retry.Max = 3
	kafkaConfig.Producer.Retry.Backoff = 100 * time.Millisecond

	// 消费者配置
	kafkaConfig.Consumer.Return.Errors = true
}

// 初始化Kafka生产者
func initKafkaProducer(brokers []string) error {
	producer, err := sarama.NewSyncProducer(brokers, kafkaConfig)
	if err != nil {
		return err
	}
	KafkaProducer = producer
	return nil
}

// 初始化Kafka消费者
func initKafkaConsumer(brokers []string, topic string) error {
	consumer, err := sarama.NewConsumer(brokers, kafkaConfig)
	if err != nil {
		return err
	}
	// 将消费者赋值给全局变量
	KafkaConsumer = consumer

	// 订阅主题
	partitionConsumer, err := KafkaConsumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	// 启动一个goroutine处理消息
	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				fmt.Printf("Received message: %s\n", string(msg.Value))
				// 在这里添加处理消息的逻辑
				address := string(msg.Value)
				fmt.Println("已经消费消息:", address)
				// 更新数据库中的领取状态
				err := DB.Model(&KafkaTest{}).Where("address = ?", address).Update("status", 1).Error
				if err != nil {
					log.Printf("更新数据库失败: %v", err)
					continue
				}

			case err := <-partitionConsumer.Errors():
				log.Printf("Received consumer error: %s\n", err.Error())
			}
		}
	}()

	return nil
}

func main() {
	var DBerr error
	// 连接 MySQL 数据库
	DB, DBerr = gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True")
	if DBerr != nil {
		log.Fatalf("failed to connect database: %v", DBerr)
	}

	// 初始化 Kafka 生产者和消费者
	brokers := []string{"localhost:9092"}
	if err := initKafkaProducer(brokers); err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}

	// 定义分区数和指定主题
	if err := initKafkaConsumer(brokers, "test"); err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}

	// 创建 Hertz 服务器
	h := server.Default(server.WithHostPorts("127.0.0.1:7071"))

	// 启动服务器
	h.Spin()
}
