// internal/monitor/service/nft_service.go
package service

import (
	"claimask/internal/monitor/dao"
	"claimask/internal/monitor/model/po"
	"claimask/pkg/dogechain"
	"errors"
	"sync"
)

// NFTDecoder 解码NFT元数据
type NFTDecoder struct {
	monitorAddr string
}

// NewNFTDecoder 创建NFT解码器
func NewNFTDecoder() *NFTDecoder {
	return &NFTDecoder{
		monitorAddr: "DTcuJ6N5QEoQUygTv8CnKzn3DUS7KhaDR2", // 默认监控地址
	}
}

// GetGTID 从交易Hash获取全局交易ID
func (d *NFTDecoder) GetGTID(txHash string) (string, error) {
	// 简化版实现，实际应调用区块链节点API获取交易详情并解析
	if txHash == "" {
		return "", errors.New("empty transaction hash")
	}

	// 仅作为演示，实际应该解析交易脚本获取GTID
	return "nft:" + txHash, nil
}

// NFTService NFT服务
type NFTService struct {
	decoder     *NFTDecoder
	dao         dao.NFTDao
	taxRate     float64
	cache       sync.Map
	monitorAddr string
}

// NewNFTService 创建NFT服务
func NewNFTService(dao dao.NFTDao, tax float64) *NFTService {
	return &NFTService{
		decoder:     NewNFTDecoder(),
		dao:         dao,
		taxRate:     tax,
		monitorAddr: "DTcuJ6N5QEoQUygTv8CnKzn3DUS7KhaDR2", // 默认监控地址
	}
}

// ProcessTransfer 处理NFT转移交易
func (s *NFTService) ProcessTransfer(tx *dogechain.TxDetail) error {
	gtid, err := s.decoder.GetGTID(tx.Hash)
	if err != nil || gtid == "" {
		return err
	}

	nftID, ok := s.cache.Load(gtid)
	if !ok {
		return nil
	}

	var taxAmt, total int64
	inputs := make(map[string]struct{})

	for _, in := range tx.Vin {
		for _, addr := range in.Addresses {
			inputs[addr] = struct{}{}
		}
	}

	var owner string
	for _, out := range tx.Vout {
		// 检查是否为NFT标记输出
		if out.Value == 100000 {
			if len(out.ScriptPubKey.Addresses) > 0 {
				owner = out.ScriptPubKey.Addresses[0]
			}
		}

		// 检查是否为税收
		if len(out.ScriptPubKey.Addresses) > 0 {
			addr := out.ScriptPubKey.Addresses[0]
			if _, exists := inputs[addr]; !exists {
				if addr == s.monitorAddr {
					taxAmt += int64(out.Value * 100000000) // 转换为ELON单位
				}
				total += int64(out.Value * 100000000)
			}
		}
	}

	taxStatus := 0
	if total > 0 && float64(taxAmt)/float64(total) >= s.taxRate {
		taxStatus = 1
	}

	// 创建NFT更新对象
	nftPO := &po.NFTPO{
		NFTID:        nftID.(string),
		UtxoHash:     tx.Hash,
		OwnerAddress: owner,
		TaxStatus:    taxStatus,
		TxAmt:        total - taxAmt,
	}

	// 更新NFT状态
	return s.dao.UpdateNFTStatus(nftPO)
}
