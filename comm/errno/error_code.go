package errno

// 错误码定义
const (
	SuccessCode               = 0
	NFTTransferError          = 5001
	UTXOInsufficientError     = 5002
	RPCConnectionError        = 5003
	TransactionBroadcastError = 5004
)

var codeMsg = map[int]string{
	SuccessCode:               "成功",
	NFTTransferError:          "NFT转移失败",
	UTXOInsufficientError:     "UTXO余额不足",
	RPCConnectionError:        "区块链节点连接失败",
	TransactionBroadcastError: "交易广播失败",
}

func GetMsg(code int) string {
	return codeMsg[code]
}
