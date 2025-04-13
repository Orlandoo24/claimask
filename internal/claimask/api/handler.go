package api

import (
	"claimask/comm/response"
	"claimask/internal/claimask/model/dto"
	"claimask/internal/claimask/service"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ClaimAPI 领取核心服务
type ClaimAPI struct {
	ClaimService service.ClaimService
}

// NewClaimAPI 创建ClaimAPI实例
func NewClaimAPI(claimService service.ClaimService) *ClaimAPI {
	return &ClaimAPI{ClaimService: claimService}
}

// Claim 处理奖品领取请求
func (api *ClaimAPI) Claim(ctx *gin.Context) {
	var param dto.ClaimParam
	if err := ctx.ShouldBindJSON(&param); err != nil {
		response.FailWithMessage(ctx, response.ERROR, "参数绑定失败: "+err.Error())
		return
	}

	// 领取奖品
	if err := api.ClaimService.ClaimPrizeV2(); err != nil {
		response.FailWithMessage(ctx, response.ERROR, "奖品数量减少失败: "+err.Error())
		return
	}

	// 创建订单
	if err := api.ClaimService.CreateOrder(param.Address); err != nil {
		response.FailWithMessage(ctx, response.ERROR, "订单创建失败: "+err.Error())
		return
	}

	response.OkWithData(ctx, param)
}

// Query 处理奖品数量查询请求
func (api *ClaimAPI) Query(ctx *gin.Context) {
	prizes, err := api.ClaimService.QueryPrizes()
	if err != nil {
		response.FailWithMessage(ctx, response.ERROR, "无法获取数量: "+err.Error())
		return
	}

	response.OkWithData(ctx, gin.H{"prizes": prizes})
}

// Init 处理奖品数量初始化请求
func (api *ClaimAPI) Init(ctx *gin.Context) {
	quantityStr := ctx.Param("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		response.FailWithMessage(ctx, response.ERROR, "无效的数量: "+err.Error())
		return
	}

	api.ClaimService.InitPrizes(quantity)
	response.OkWithMessage(ctx, fmt.Sprintf("奖品数量已重置为 %d", quantity))
}
