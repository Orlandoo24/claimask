package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// GOWORK=off go run main.go

func main() {
	// 硬编码Kafka连接参数
	brokers := "9.134.132.205:19092"
	topic := "test-topic"
	username := "admin"
	password := "123456"
	mechanism := "SCRAM-SHA-256"
	groupID := "my-go-consumer-group"

	// 生产者配置
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanism":    mechanism,
		"sasl.username":     username,
		"sasl.password":     password,
	})
	if err != nil {
		log.Fatal("生产者创建失败: ", err)
	}
	defer producer.Close()

	// 发送测试消息
	message := "Hello Kafka " + time.Now().Format(time.RFC3339)
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)

	if err != nil {
		log.Fatal("消息发送失败: ", err)
	}

	// 等待消息发送完成
	producer.Flush(15 * 1000)
	fmt.Println("消息已发送到Topic:", topic, "内容:", message)

	// 消费者配置
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"security.protocol":  "SASL_PLAINTEXT",
		"sasl.mechanism":     mechanism,
		"sasl.username":      username,
		"sasl.password":      password,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "true",
	})

	if err != nil {
		log.Fatal("消费者创建失败: ", err)
	}
	defer consumer.Close()

	// 订阅主题
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatal("订阅主题失败: ", err)
	}
	fmt.Println("已订阅主题:", topic)

	// 捕获终止信号
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// 接收消息的计时器
	timeout := 30 * time.Second
	timeoutChan := time.After(timeout)

	// 循环接收消息
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("捕获到信号 %v: 终止程序\n", sig)
			run = false
		case <-timeoutChan:
			fmt.Printf("超时 %v 秒: 终止程序\n", timeout.Seconds())
			run = false
		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("收到消息: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s\n",
					*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset,
					string(e.Key), string(e.Value))
				// 如果接收到刚才发送的消息，退出循环
				if string(e.Value) == message {
					fmt.Println("已接收到刚才发送的消息，程序将退出")
					run = false
				}
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

	fmt.Println("程序已正常退出")
}
