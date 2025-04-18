package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/viper"
)

// Config 配置结构
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

	fmt.Println("开始创建Kafka消费者...")
	fmt.Printf("连接到服务器: %s\n", cfg.Kafka.Brokers)
	fmt.Printf("主题: %s\n", cfg.Kafka.Topic)

	// 创建消费者
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Brokers,
		"group.id":          "test-consumer-group",
		"auto.offset.reset": "earliest",
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanism":    cfg.Kafka.SASL.Mechanism,
		"sasl.username":     cfg.Kafka.SASL.Username,
		"sasl.password":     cfg.Kafka.SASL.Password,
	})

	if err != nil {
		log.Fatalf("消费者创建失败: %v", err)
	}
	defer consumer.Close()

	// 订阅主题
	err = consumer.SubscribeTopics([]string{cfg.Kafka.Topic}, nil)
	if err != nil {
		log.Fatalf("主题订阅失败: %v", err)
	}

	fmt.Printf("成功订阅主题: %s, 等待消息...\n", cfg.Kafka.Topic)

	// 捕获中断信号
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("捕获到信号 %v: 终止消费者\n", sig)
			run = false
		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("收到消息: %s\n", string(e.Value))
				fmt.Printf("主题分区: %v, 偏移量: %v\n", e.TopicPartition.Partition, e.TopicPartition.Offset)
			case kafka.Error:
				fmt.Printf("错误: %v\n", e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("忽略事件: %v\n", e)
			}
		}
	}

	fmt.Println("消费者已停止")
}
