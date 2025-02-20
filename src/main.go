package main

import (
	"claimask/src/claim"
	"claimask/src/comm/initialize"
	"log"
)

func main() {
	// 初始化日志
	initialize.InitLogger()

	// 初始化数据库
	db, err := initialize.InitDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 初始化 Redis
	rd := initialize.InitRedis()

	// 初始化 DAO 和 Service
	orderDAO := claim.NewOrderDAO(db)
	orderService := claim.NewOrderService(orderDAO, rd)

	// 初始化 API
	claimAPI := claim.NewClaimAPI(orderService)

	// 设置路由
	r := claim.SetupRouter(claimAPI)

	// 启动服务器
	r.Run(":8870")
}
