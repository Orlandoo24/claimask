package initialize

import (
	"claimask/pkg/dogechain"
	"fmt"

	"go.uber.org/zap"
)

func InitDogecoinRPC(ip string, port int, user, password string) *dogechain.RPCClient {
	// 构造完整的RPC端点URL
	endpoint := fmt.Sprintf("http://%s:%d", ip, port)

	// 创建RPC客户端
	client := dogechain.NewRPCClient(endpoint, user, password)

	// 测试连接
	zap.L().Info("初始化Dogecoin RPC客户端",
		zap.String("endpoint", endpoint))

	// 这里可以添加连接测试，但为了简化，我们只打印警告不做测试
	zap.L().Warn("注意：未验证Dogecoin RPC连接可用性，如果连接失败可能在使用时报错",
		zap.String("endpoint", endpoint))

	return client
}
