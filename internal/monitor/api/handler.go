package api

import (
	"claimask/internal/monitor/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	monitorSvc service.MonitorService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(svc service.MonitorService) *PaymentHandler {
	return &PaymentHandler{monitorSvc: svc}
}

// HandlePaymentCallback 处理支付回调
func (h *PaymentHandler) HandlePaymentCallback(c *gin.Context) {
	var req struct {
		UserURL   string `json:"userUrl"`
		PayAmount int64  `json:"payAmt"`
		TxID      string `json:"txId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("支付回调参数绑定失败", zap.Error(err))
		c.JSON(400, gin.H{"code": 4001, "msg": "参数格式错误"})
		return
	}

	if err := h.monitorSvc.ProcessPayment(c.Request.Context(), req.UserURL, req.PayAmount, req.TxID); err != nil {
		zap.L().Warn("支付处理失败", zap.String("txid", req.TxID), zap.Error(err))
		c.JSON(500, gin.H{"code": 5001, "msg": "支付处理失败"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "处理成功"})
}

// GetNFTStatus 获取NFT状态
func (h *PaymentHandler) GetNFTStatus(c *gin.Context) {
	txid := c.Param("txid")
	if txid == "" {
		c.JSON(400, gin.H{"code": 4001, "msg": "参数错误"})
		return
	}

	status, err := h.monitorSvc.GetNFTStatus(c.Request.Context(), txid)
	if err != nil {
		zap.L().Warn("获取NFT状态失败", zap.String("txid", txid), zap.Error(err))
		c.JSON(500, gin.H{"code": 5001, "msg": "获取NFT状态失败"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": status})
}
