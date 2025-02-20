package initialize

import (
	"claimask/src/claim/model"
	"log"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// InitDB 初始化数据库连接并迁移表结构
func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", "root:123@jiaru@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True")
	if err != nil {
		return nil, err
	}

	// 自动迁移表结构
	db.AutoMigrate(&model.Order{})

	return db, nil
}

// InitRedis 初始化Redis连接
func InitRedis() *redis.Client {
	rd := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// 在 Redis 中设置奖品数量
	rd.Set("prizes", 5, 0)

	return rd
}

// InitLogger 初始化日志
func InitLogger() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
