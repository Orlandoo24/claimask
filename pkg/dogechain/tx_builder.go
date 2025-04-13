package dogechain

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func TransferElonUseUtxo(sender, privKey, receiver string, utxos []UTXO) (string, error) {
	tx := wire.NewMsgTx(wire.TxVersion)

	// 添加UTXO输入
	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(utxo.TxHash)
		if err != nil {
			return "", fmt.Errorf("invalid hash: %w", err)
		}
		outPoint := wire.NewOutPoint(hash, utxo.Index)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
	}

	// 构建输出
	receiverAddr, err := btcutil.DecodeAddress(receiver, &chaincfg.MainNetParams)
	if err != nil {
		return "", fmt.Errorf("invalid receiver address: %w", err)
	}
	pkScript, err := txscript.PayToAddrScript(receiverAddr)
	if err != nil {
		return "", fmt.Errorf("failed to create pkScript: %w", err)
	}
	tx.AddTxOut(wire.NewTxOut(utxos[0].Value, pkScript))

	// 签名
	// 简化实现，仅仅返回序列化的交易
	// 注意：实际项目中应该实现完整的签名逻辑

	// 返回十六进制格式交易
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	tx.Serialize(buf)
	return hex.EncodeToString(buf.Bytes()), nil
}

type UTXO struct {
	TxHash string
	Index  uint32
	Value  int64
}
