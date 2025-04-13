package constant

const (
	// 单位转换系数（1 DOGE = 100,000,000 ELON）
	DOGE_TO_ELON = 100000000
	ELON_TO_DOGE = 1.0 / DOGE_TO_ELON

	// NFT标识符
	NFT_PREFIX       = "ord"
	NFT_CONTENT_TYPE = "text/plain;charset=utf-8"

	// 交易参数
	DEFAULT_FEE_RATE   = 50000  // 默认费率（ELON/byte）
	MINIMUM_UTXO_VALUE = 100000 // 最小UTXO值（ELON）
)
