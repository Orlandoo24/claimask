package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Shopify/sarama"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// 定义一个结构体来接收POST请求中的JSON数据
type Address struct {
	Address string `json:"address"`
}

func main() {

	fmt.Println("producer starting server on port 7070")

	// 连接到MySQL数据库
	// "root:@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True"是数据库的连接字符串
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True")
	if err != nil {
		log.Fatal(err) // 如果连接失败，记录错误并退出程序
	}
	defer db.Close() // 确保在main函数结束时关闭数据库连接

	// 创建Kafka生产者的配置
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	// 使用配置创建Kafka生产者
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err) // 如果创建失败，记录错误并退出程序
	}
	defer producer.Close() // 确保在main函数结束时关闭Kafka生产者

	// 定义一个HTTP处理函数，它会在收到POST请求时被调用
	http.HandleFunc("/producer", func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法是否为POST，如果不是，返回一个错误
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var a Address
		// 从请求的Body中解析JSON数据，并将其存储在一个Address结构体中
		err := json.NewDecoder(r.Body).Decode(&a)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 打印入参
		fmt.Println("msg:", a.Address)

		// 将数据插入到数据库中
		_, err = db.Exec("INSERT INTO kafka_test (address, status) VALUES (?, 0)", a.Address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 创建一个Kafka消息
		msg := &sarama.ProducerMessage{
			Topic: "test",
			Value: sarama.StringEncoder(a.Address),
		}

		// 将消息发送到Kafka
		_, _, err = producer.SendMessage(msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 返回一个成功的响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// 在端口7070上启动HTTP服务器
	fmt.Println("Starting server on port 7070")
	log.Fatal(http.ListenAndServe("127.0.0.1:7070", nil))
}
