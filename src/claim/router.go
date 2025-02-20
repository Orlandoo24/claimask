package claim

import (
	"claimask/src/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(api *ClaimAPI) *gin.Engine {
	r := gin.Default()

	// 使用跨域中间件
	r.Use(middleware.CorsMiddleware())

	// 定义名额领取接口
	r.POST("/claim", api.Claim)

	// 定义数量查询接口
	r.GET("/query", api.Query)

	// 新增接口 "/initialize" 实现奖品数量重置功能
	r.GET("/initialize/:quantity", api.Init)

	return r
}
