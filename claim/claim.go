package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/hertz-contrib/cors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

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

func main() {
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

	// 创建 Hertz 服务器
	h := server.Default(server.WithHostPorts("127.0.0.1:8870"))

	h.Use(cors.New(cors.Config{
		// 允许跨源访问的 origin 列表
		AllowOrigins: []string{"*"},
		// 允许客户端跨源访问所使用的 HTTP 方法列表
		AllowMethods: []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		// 允许使用的头信息字段列表
		AllowHeaders: []string{"Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma"},
		// 允许暴露给客户端的响应头列表
		ExposeHeaders: []string{"Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar"},
		// 允许客户端请求携带用户凭证
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 定义领奖接口
	h.POST("/claim", func(c context.Context, ctx *app.RequestContext) {
		// 定义请求参数结构体
		type ClaimParam struct {
			Address string `json:"address"`
		}

		// 创建参数实例
		var param ClaimParam
		// 绑定请求参数到结构体
		bindErr := ctx.Bind(&param)
		// 打印参数地址
		fmt.Printf("Param address: %s\n", param.Address)

		// 如果绑定出错，返回错误信息
		if bindErr != nil {
			ctx.String(consts.StatusBadRequest, "bind error: %s", bindErr.Error())
			return
		}

		// 从 Redis 中获取奖品数量
		var prizes int
		for {
			err := RD.Watch(func(tx *redis.Tx) error {
				var err error
				prizes, err = tx.Get("prizes").Int()
				if err != nil {
					return err
				}

				// 如果奖品数量大于 0，则递减奖品数量
				if prizes > 0 {
					_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
						// 在 redis 减少奖品数量
						pipe.Decr("prizes")
						return nil
					})
					if err != nil {
						return err
					}
					return nil
				}
				return errors.New("奖品已经领完了")
			}, "prizes")

			if err == nil {
				break
			} else if err == redis.TxFailedErr {
				continue // Retry
			} else {
				ctx.String(consts.StatusInternalServerError, "奖品数量减少失败")
				return
			}
		}

		// 创建订单
		order := &Order{
			OrderID:    generateOrderID(),
			Address:    param.Address,
			Json1:      `{"key": "value"}`,
			InsertTime: time.Now(), // 分配当前时间戳值
			UpdateTime: time.Now(), // 分配当前时间戳值
		}

		// 在数据库中创建订单
		dbErr := DB.Table("order_id").Create(order).Error
		if dbErr != nil {
			ctx.String(consts.StatusInternalServerError, "订单创建失败")
			return
		}

		// 返回成功信息
		ctx.JSON(consts.StatusOK, utils.H{"param": param})
	})

	// 新增接口 "/init" 实现奖品数量重置功能
	h.GET("/init", func(c context.Context, ctx *app.RequestContext) {
		// 重置奖品数量为 10
		RD.Set("prizes", 5, 0)
		ctx.String(consts.StatusOK, "奖品数量已重置为 5")
	})

	// 启动服务器
	h.Spin()

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
