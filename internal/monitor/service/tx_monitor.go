// internal/monitor/service/tx_monitor.go
package service

import (
	"context"
	"sync"
	"time"

	"claimask/pkg/dogechain"

	"go.uber.org/zap"
)

// MonitorConfig 监控配置
type MonitorConfig struct {
	WalletGroups      []string      // 监控的钱包地址组
	BlockPollInterval time.Duration // 区块轮询间隔
	WebsocketEndpoint string        // WebSocket端点
}

// PaymentService 支付服务
type PaymentService struct {
	callbackURL string   // 支付回调URL
	processors  []string // 支付处理器列表
}

// Process 处理支付
func (p *PaymentService) Process(ctx context.Context, from string, amount int64, txHash string) error {
	// 简化版支付处理逻辑
	zap.L().Info("处理支付交易",
		zap.String("from", from),
		zap.Int64("amount", amount),
		zap.String("txHash", txHash))

	// 实际项目中应该实现支付处理逻辑
	return nil
}

// TxMonitor 交易监控器
type TxMonitor struct {
	rpcClient     *dogechain.RPCClient
	nftSvc        *NFTService
	paymentSvc    *PaymentService
	config        *MonitorConfig
	ctx           context.Context
	cancel        context.CancelFunc
	nftMap        sync.Map
	monitorAddrs  []string
	lastBlockHash string
}

// NewTxMonitor 创建交易监控器
func NewTxMonitor(rpc *dogechain.RPCClient, cfg *MonitorConfig) *TxMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &TxMonitor{
		rpcClient:    rpc,
		config:       cfg,
		ctx:          ctx,
		cancel:       cancel,
		monitorAddrs: cfg.WalletGroups,
		paymentSvc:   &PaymentService{callbackURL: "http://localhost/callback"},
	}
}

// isNFTOperation 判断是否为NFT操作
func (m *TxMonitor) isNFTOperation(tx *dogechain.TxDetail) bool {
	for _, out := range tx.Vout {
		// NFT标志：100,000 ELON (0.001 DOGE)
		if out.Value == 0.001 {
			return true
		}
	}
	return false
}

// handlePayment 处理支付
func (m *TxMonitor) handlePayment(tx *dogechain.TxDetail) error {
	var inOur, outOur bool
	var senderAddr string

	// 检查输入是否包含我们的地址
	for _, in := range tx.Vin {
		for _, addr := range in.Addresses {
			if contains(m.monitorAddrs, addr) {
				inOur = true
				break
			}
		}
		if inOur {
			break
		}
		// 记录第一个输入地址作为发送者
		if len(in.Addresses) > 0 && senderAddr == "" {
			senderAddr = in.Addresses[0]
		}
	}

	// 如果输入不是我们的地址，检查输出
	if !inOur {
		var total int64
		for _, out := range tx.Vout {
			for _, addr := range out.ScriptPubKey.Addresses {
				if contains(m.monitorAddrs, addr) {
					outOur = true
					total += int64(out.Value * 100000000) // 转换为ELON单位
				}
			}
		}

		// 如果输出包含我们的地址，处理为支付
		if outOur {
			return m.paymentSvc.Process(context.Background(),
				senderAddr,
				total,
				tx.Hash,
			)
		}
	}
	return nil
}

// StartDualMonitor 启动双通道监控
func (m *TxMonitor) StartDualMonitor() {
	go m.processNodeWebsocket()
	go m.pollBlockExplorer()
}

// processNodeWebsocket 处理节点WebSocket消息
func (m *TxMonitor) processNodeWebsocket() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.checkNewBlocks(); err != nil {
				zap.L().Warn("区块处理错误", zap.Error(err))
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// pollBlockExplorer 轮询区块浏览器
func (m *TxMonitor) pollBlockExplorer() {
	ticker := time.NewTicker(m.config.BlockPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 实现区块浏览器API调用
			zap.L().Debug("轮询区块浏览器")
		case <-m.ctx.Done():
			return
		}
	}
}

// checkNewBlocks 检查新区块
func (m *TxMonitor) checkNewBlocks() error {
	// 实现检查新区块逻辑
	zap.L().Debug("检查新区块")
	return nil
}

// contains 辅助函数：检查数组是否包含指定值
func contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
