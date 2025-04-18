package main

import (
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/viper"
)

type Config struct {
	Kafka struct {
		Brokers string `mapstructure:"brokers"`
		Topic   string `mapstructure:"topic"`
		SASL    struct {
			Username  string `mapstructure:"username"`
			Password  string `mapstructure:"password"`
			Mechanism string `mapstructure:"mechanism"`
		} `mapstructure:"sasl"`
	} `mapstructure:"kafka"`
}

func main() {
	// 读取配置
	v := viper.New()
	v.SetConfigFile("../kafka.yaml")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("配置文件读取失败: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("配置解析失败: %v", err)
	}

	fmt.Println("开始创建Kafka生产者...")
	fmt.Printf("连接到服务器: %s\n", cfg.Kafka.Brokers)
	fmt.Printf("主题: %s\n", cfg.Kafka.Topic)

	// 连接Kafka
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Brokers,
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanism":    cfg.Kafka.SASL.Mechanism,
		"sasl.username":     cfg.Kafka.SASL.Username,
		"sasl.password":     cfg.Kafka.SASL.Password,
	})
	if err != nil {
		log.Fatalf("Kafka连接失败: %v", err)
	}
	defer producer.Close()

	// 发送测试消息
	topic := cfg.Kafka.Topic
	deliveryChan := make(chan kafka.Event)

	// 使用当前时间戳创建消息内容
	currentTime := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf("Hello Kafka - %s", currentTime)

	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, deliveryChan)

	if err != nil {
		log.Fatalf("消息发送失败: %v", err)
	}

	// 等待消息确认
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Fatalf("消息发送失败: %v", m.TopicPartition.Error)
	}

	fmt.Printf("消息 '%s' 已发送到 %s 的分区 %d\n", message, *m.TopicPartition.Topic, m.TopicPartition.Partition)
}
