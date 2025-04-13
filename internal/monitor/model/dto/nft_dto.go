package dto

type NFTUpdateRequest struct {
	NFTID        string `json:"nft_id" binding:"required"`
	UtxoHash     string `json:"nft_utxo" binding:"required"`
	OwnerAddress string `json:"owner_address"`
	TaxStatus    int    `json:"tax_status"` // 0-未缴税 1-已缴税
	TxAmt        int64  `json:"tx_amt"`     // 交易金额（ELON）
}
