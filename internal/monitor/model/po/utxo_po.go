package po

type UTXOPO struct {
	ID      uint   `gorm:"primaryKey"`
	Address string `gorm:"index"`
	TxHash  string `gorm:"size:64;uniqueIndex"`
	Index   uint32 `gorm:"index"`
	Value   int64  // 单位：ELON
	Spent   bool   `gorm:"default:false"`
}

type NFTPO struct {
	NFTID        string `gorm:"primaryKey;size:64"` // 唯一标识符
	UtxoHash     string `gorm:"size:64;uniqueIndex"`
	OwnerAddress string `gorm:"size:34"`   // 当前所有者
	TaxStatus    int    `gorm:"default:0"` // 0-未缴税 1-已缴税
	TxAmt        int64  `gorm:"default:0"` // 交易金额（ELON）
	GTID         string `gorm:"size:128"`  // 全局交易标识
}

type WalletGroupPO struct {
	GroupID      int    `gorm:"primaryKey"`
	ReceiveAddr  string `gorm:"size:34;uniqueIndex"` // Dogecoin地址
	PrivateKey   string `gorm:"type:text"`           // 加密存储
	CurrentUTXO  string `gorm:"size:64"`             // 当前使用的UTXO
	LastSyncTime int64  // 最后同步时间戳
}
