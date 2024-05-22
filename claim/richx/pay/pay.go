package main

import (
	"fmt"
)

// 定义各个组件
type PayTradePlatformApiBizServiceImpl struct{}
type TradeOrderPayFacadeAdapterImpl struct{}
type TradeOrderBizServiceImpl struct{}
type Database struct{}

// 下支付订单
func (t *TradeOrderBizServiceImpl) CreateOrder() {
	fmt.Println("下支付订单")
}

// SDK初始化
func (p *PayTradePlatformApiBizServiceImpl) InitSDK() {
	fmt.Println("SDK初始化")
}

// 根据订单号查询订单
func (t *TradeOrderBizServiceImpl) QueryOrder() {
	fmt.Println("根据订单号查询订单")
}

// 下单到账查询
func (t *TradeOrderBizServiceImpl) CheckPayment() {
	fmt.Println("下单到账查询")
}

// 从数据库获取支付结果
func (d *Database) GetPaymentResult() {
	fmt.Println("从数据库获取支付结果")
}

func main() {
	var payService PayTradePlatformApiBizServiceImpl
	var orderService TradeOrderBizServiceImpl
	var database Database

	orderService.CreateOrder()
	payService.InitSDK()
	orderService.QueryOrder()
	orderService.CheckPayment()
	database.GetPaymentResult()
}
