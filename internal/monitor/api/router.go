package api

import (
	"astro-orderx/internal/monitor/service"
	"astro-orderx/pkg/dogechain"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func RegisterRoutes(router *gin.Engine, rpcClient interface{}, redisClient interface{}) {
	monitorSvc := service.NewMonitorService(
		service.NewTxMonitor(rpcClient.(*dogechain.RPCClient), &service.MonitorConfig{
			WalletGroups:      []string{"DTcuJ6N5QEoQUygTv8CnKzn3DUS7KhaDR2"},
			BlockPollInterval: 60 * time.Second,
			WebsocketEndpoint: "wss://ws.dogechain.info/",
		}),
		service.NewQueueManager(redisClient.(*redis.Client)),
	)

	handler := NewPaymentHandler(monitorSvc)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/pay-callback", handler.HandlePaymentCallback)
		v1.GET("/nft-status/:txid", handler.GetNFTStatus)
	}
}
