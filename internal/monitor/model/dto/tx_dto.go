package dto

type TxInput struct {
	Address string `json:"address"`
	Value   int64  `json:"value"`
	TxHash  string `json:"tx_hash"`
	Index   uint32 `json:"index"`
}

type TxOutput struct {
	Address string `json:"address"`
	Value   int64  `json:"value"`
}

type TxDTO struct {
	Hash    string     `json:"hash"`
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}
