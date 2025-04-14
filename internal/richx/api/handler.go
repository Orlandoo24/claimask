package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"claimask/internal/richx/model"
	"claimask/internal/richx/service"
)

// ClaimHandler 处理富豪奖励相关请求
type ClaimHandler struct {
	claimService *service.ClaimService
}

// NewClaimHandler 创建ClaimHandler实例
func NewClaimHandler(claimService *service.ClaimService) *ClaimHandler {
	return &ClaimHandler{
		claimService: claimService,
	}
}

// SetupRouter 配置路由
func (h *ClaimHandler) SetupRouter(r *gin.Engine) {
	// POST /rich/claim 处理领取请求
	r.POST("/rich/claim", func(c *gin.Context) {
		var req model.ClaimRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
			})
			return
		}

		resp, err := h.claimService.Claim(c, req.Address)
		if err != nil {
			// 错误已在服务层处理，直接返回
			c.JSON(http.StatusOK, resp)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	// GET /rich/claim 查询当前奖励状态
	r.GET("/rich/claim", func(c *gin.Context) {
		address := c.Query("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "缺少地址参数",
			})
			return
		}

		canClaim, amount, err := h.claimService.CanClaim(address)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "查询失败，请稍后再试",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"canClaim": canClaim,
				"amount":   amount.String(),
			},
			"message": "查询成功",
		})
	})
}
