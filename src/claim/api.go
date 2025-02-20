package claim

import (
	"claimask/src/claim/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClaimAPI struct {
	OrderService OrderService
}

func NewClaimAPI(orderService OrderService) *ClaimAPI {
	return &ClaimAPI{OrderService: orderService}
}

func (api *ClaimAPI) Claim(ctx *gin.Context) {
	var param model.ClaimParam
	if err := ctx.ShouldBindJSON(&param); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := api.OrderService.ClaimPrize(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "奖品数量减少失败"})
		return
	}

	if err := api.OrderService.CreateOrder(param.Address); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "订单创建失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"param": param})
}

func (api *ClaimAPI) Query(ctx *gin.Context) {
	prizes, err := api.OrderService.QueryPrizes()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取数量"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"prizes": prizes})
}

func (api *ClaimAPI) Init(ctx *gin.Context) {
	quantityStr := ctx.Param("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的数量"})
		return
	}

	api.OrderService.InitPrizes(quantity)
	ctx.String(http.StatusOK, fmt.Sprintf("奖品数量已重置为 %d", quantity))
}
