package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterClaimRoutes 设置领取相关路由
func RegisterClaimRoutes(r *gin.RouterGroup, api *ClaimAPI) {
	// 领取奖品相关路由
	claimGroup := r.Group("/claim")
	{
		// 定义名额领取接口
		claimGroup.POST("", api.Claim)

		// 定义数量查询接口
		claimGroup.GET("/query", api.Query)

		// 新增接口 "/initialize" 实现奖品数量重置功能
		claimGroup.GET("/initialize/:quantity", api.Init)
	}
}
