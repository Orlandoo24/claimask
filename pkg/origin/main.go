package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

/**
 *  model 层
 */

// 定义订单结构体
type Order struct {
	ID         uint      `gorm:"primary_key"`
	OrderID    uint64    `gorm:"column:order_id"`
	Address    string    `gorm:"column:address"`
	Json1      string    `gorm:"column:json1"`
	InsertTime time.Time `gorm:"column:insert_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

// 定义全局变量
var (
	DB *gorm.DB
	RD *redis.Client
)

/**
 *  dao
 */

// 定义请求参数结构体
type ClaimParam struct {
	Address string `json:"address"`
}

// 全局变量声明
var globalClaimParam = ClaimParam{}

func main() {

	/***
	 * 初始化包 initialize
	 */
	var err error
	// 连接 MySQL 数据库
	DB, err = gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/faker?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic("failed to connect database")
	}

	// 指定表名为 "order_id"
	DB.Table("order_id").AutoMigrate(&Order{})

	// 连接 Redis
	RD = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// 在 Redis 中设置奖品数量
	RD.Set("prizes", 5, 0)

	// 创建 Gin 服务器
	r := gin.Default()

	/**
	 * 中间件包
	 */
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Length", "X-CSRF-Token", "Token", "session", "X_Requested_With", "Accept", "Origin", "Host", "Connection", "Accept-Encoding", "Accept-Language", "DNT", "X-CustomHeader", "Keep-Alive", "User-Agent", "If-Modified-Since", "Cache-Control", "Content-Type", "Pragma"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Cache-Control", "Content-Language", "Content-Type", "Expires", "Last-Modified", "Pragma", "FooBar"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	/**
	 * router 包
	 */

	// 定义名额领取接口
	r.POST("/claim", func(ctx *gin.Context) {
		var param ClaimParam
		if err := ctx.ShouldBindJSON(&param); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Printf("Param address: %s\n", param.Address)

		if err := ClaimPrize(RD); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "奖品数量减少失败"})
			return
		}

		if err := CreateOrder(param); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "订单创建失败"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"param": param})
	})

	// 定义数量查询接口
	r.GET("/query", func(ctx *gin.Context) {
		prizes, err := RD.Get("prizes").Int()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取数量"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"prizes": prizes})
	})

	// 新增接口 "/initialize" 实现奖品数量重置功能
	r.GET("/initialize/:quantity", func(ctx *gin.Context) {
		quantityStr := ctx.Param("quantity")
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的数量"})
			return
		}

		RD.Set("prizes", quantity, 0)
		ctx.String(http.StatusOK, fmt.Sprintf("奖品数量已重置为 %d", quantity))
	})

	// 启动服务器
	r.Run(":8870")
}

/**
 * service 包
 */

// ClaimPrize 用于领取奖品的函数，传入一个 Redis 客户端 RD，返回可能的错误
func ClaimPrize(RD *redis.Client) error {
	var prizes int                 // 声明奖品数量变量
	maxRetries := 3                // 最大重试次数
	maxDuration := 5 * time.Second // 最大执行时间限制，假设为5秒

	retries := 0        // 初始化重试次数为0
	start := time.Now() // 记录开始时间

	for {
		err := RD.Watch(func(tx *redis.Tx) error {
			var err error

			// 从 Redis 中获取奖品数量
			prizes, err = tx.Get("prizes").Int()
			if err != nil {
				return err
			}

			// 如果奖品数量大于 0，则递减奖品数量
			if prizes > 0 {
				_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
					// 在 Redis 中递减奖品数量
					pipe.Decr("prizes")
					return nil
				})
				if err != nil {
					return err
				}
				return nil
			}
			return errors.New("奖品已经领完了") // 如果奖品数量为0，则返回错误信息
		}, "prizes")

		if err == nil {
			// 如果没有错误，表示奖品数量获取和递减成功，跳出循环
			break
		} else if err == redis.TxFailedErr {
			fmt.Print("当前有其他事务对 prizes 键进行了修改，事务回滚，并进行重试")

			// 如果出现 redis.TxFailedErr 错误，表示事务失败，需要重试
			retries++
			if retries >= maxRetries || time.Since(start) >= maxDuration {
				// 如果达到最大重试次数或者超过最大时间限制，退出循环并返回错误信息
				return errors.New("重试次数超过限制或执行时间超时")
			}
			continue // Retry
		} else {
			// 如果出现其他错误，返回给客户端错误信息，并结束处理
			return err
		}
	}
	return nil
}

func CreateOrder(param ClaimParam) error {
	order := &Order{
		OrderID:    generateOrderID(),
		Address:    param.Address,
		Json1:      `{"key": "value"}`,
		InsertTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 在数据库中创建订单
	dbErr := DB.Table("order_id").Create(order).Error
	if dbErr != nil {
		return dbErr
	}
	return nil
}

// 分布式 id
func generateOrderID() uint64 {
	// 创建一个新的节点（Node），用于生成雪花ID
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("无法创建雪花节点: %v", err)
	}

	// 生成一个新的雪花ID
	id := node.Generate()

	// 将 int64 类型的 ID 转换为 uint64 类型
	return uint64(id.Int64())
}
