package claim

import (
	"claimask/src/claim/model"
	"claimask/src/comm/resp"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ClaimAPI 领取核心服务
type ClaimAPI struct {
	OrderService OrderService
}

func NewClaimAPI(orderService OrderService) *ClaimAPI {
	return &ClaimAPI{OrderService: orderService}
}

func (api *ClaimAPI) Claim(ctx *gin.Context) {
	var param model.ClaimParam
	if err := ctx.ShouldBindJSON(&param); err != nil {
		response.FailWithMessage(ctx, response.ERROR, "参数绑定失败: "+err.Error())
		return
	}

	// 领取奖品
	//if err := api.OrderService.ClaimPrize(); err != nil {
	//	response.FailWithMessage(ctx, response.ERROR, "奖品数量减少失败: "+err.Error())
	//	return
	//}

	if err := api.OrderService.ClaimPrizeV2(); err != nil {
		response.FailWithMessage(ctx, response.ERROR, "奖品数量减少失败: "+err.Error())
		return
	}

	// 创建订单
	if err := api.OrderService.CreateOrder(param.Address); err != nil {
		response.FailWithMessage(ctx, response.ERROR, "订单创建失败: "+err.Error())
		return
	}

	response.OkWithData(ctx, param)
}

func (api *ClaimAPI) Query(ctx *gin.Context) {
	prizes, err := api.OrderService.QueryPrizes()
	if err != nil {
		response.FailWithMessage(ctx, response.ERROR, "无法获取数量: "+err.Error())
		return
	}

	response.OkWithData(ctx, gin.H{"prizes": prizes})
}

func (api *ClaimAPI) Init(ctx *gin.Context) {
	quantityStr := ctx.Param("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		response.FailWithMessage(ctx, response.ERROR, "无效的数量: "+err.Error())
		return
	}

	api.OrderService.InitPrizes(quantity)
	response.OkWithMessage(ctx, fmt.Sprintf("奖品数量已重置为 %d", quantity))
}
