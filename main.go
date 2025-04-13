package main

import (
	"claimask/comm/initialize"
	"claimask/comm/utils"
	claimaskAPI "claimask/internal/claimask/api"
	claimaskDao "claimask/internal/claimask/dao"
	claimaskService "claimask/internal/claimask/service"
	monitorAPI "claimask/internal/monitor/api"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 初始化配置
	viper.SetConfigFile("./conf/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("配置文件读取失败: %v", err))
	}

	// 初始化日志
	utils.InitLogger()

	// 初始化RPC连接
	rpcClient := initialize.InitDogecoinRPC(
		viper.GetString("rpc.ip"),
		viper.GetInt("rpc.port"),
		viper.GetString("rpc.user"),
		viper.GetString("rpc.password"),
	)

	// 初始化Redis
	redisClient := initialize.InitRedis(
		viper.GetString("redis.addr"),
		viper.GetString("redis.password"),
		viper.GetInt("redis.db"),
	)

	// 初始化数据库连接
	var db *gorm.DB
	// 这里假设我们有DB初始化函数，根据实际情况使用正确的函数
	var err error
	if db, err = initialize.InitDB(); err != nil {
		zap.L().Fatal("数据库初始化失败", zap.Error(err))
	}
	defer db.Close()

	// 创建Gin引擎
	router := gin.Default()

	// 注册监控服务路由 - Monitor服务
	monitorAPI.RegisterRoutes(router, rpcClient, redisClient)

	// 初始化ClaimMask相关服务
	orderDAO := claimaskDao.NewOrderDAO(db)
	claimService := claimaskService.NewClaimService(orderDAO, redisClient)
	claimAPI := claimaskAPI.NewClaimAPI(claimService)

	// 注册ClaimMask路由
	apiGroup := router.Group("/api")
	claimaskAPI.RegisterClaimRoutes(apiGroup, claimAPI)

	// 启动服务
	port := viper.GetString("server.port")
	zap.L().Info("服务启动成功", zap.String("port", port))
	if err := router.Run(":" + port); err != nil {
		zap.L().Fatal("服务启动失败", zap.Error(err))
	}
}
