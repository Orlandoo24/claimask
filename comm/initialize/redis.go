package initialize

import (
	"log"

	"github.com/go-redis/redis"
)

// InitRedis 初始化Redis连接
func InitRedis(addr, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}

	return client
}
