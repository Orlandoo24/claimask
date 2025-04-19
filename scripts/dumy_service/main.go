package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志
	logger := log.New(log.Writer(), "MockChainService: ", log.LstdFlags)

	router := gin.Default()

	router.POST("/claim", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Printf("解析请求参数失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
			})
			return
		}

		// 打印接收到的参数
		logger.Printf("接收到领取请求参数:")
		logger.Printf("钱包地址: %v", req["address"])
		logger.Printf("领取数额: %v", req["amount"])
		logger.Printf("订单号: %v", req["orderId"])
		logger.Printf("签名: %v", req["sha"])
		logger.Printf("发送方: %v", req["from"])
		logger.Printf("NFT UTXO: %v", req["nftUtxo"])

		// 打印收益发放信息
		logger.Printf("已向地址 %v 发放收益 %v，订单号: %v", req["address"], req["amount"], req["orderId"])

		// 返回成功响应
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "收益发放成功",
			"data": gin.H{
				"txid": "0x" + time.Now().Format("20060102150405") + "000000000000000000000000",
			},
		})
	})

	// 启动服务
	logger.Printf("模拟链上服务启动在 http://localhost:7777")
	if err := router.Run(":7777"); err != nil {
		logger.Fatalf("启动模拟链上服务失败: %v", err)
	}
}
