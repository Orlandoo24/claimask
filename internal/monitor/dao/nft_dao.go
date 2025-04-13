package dao

import (
	"astro-orderx/internal/monitor/model/po"

	"gorm.io/gorm"
)

type NFTDao interface {
	UpdateNFTStatus(nft *po.NFTPO) error
	GetNFTByUTXO(utxoHash string) (*po.NFTPO, error)
}

type NFTDaoImpl struct {
	db *gorm.DB
}

func NewNFTDao(db *gorm.DB) NFTDao {
	return &NFTDaoImpl{db: db}
}

func (d *NFTDaoImpl) UpdateNFTStatus(nft *po.NFTPO) error {
	return d.db.Where("utxo_hash = ?", nft.UtxoHash).
		Assign(map[string]interface{}{
			"owner_address": nft.OwnerAddress,
			"tax_status":    nft.TaxStatus,
			"tx_amt":        nft.TxAmt,
		}).FirstOrCreate(nft).Error
}

func (d *NFTDaoImpl) GetNFTByUTXO(utxoHash string) (*po.NFTPO, error) {
	var nft po.NFTPO
	result := d.db.Where("utxo_hash = ?", utxoHash).First(&nft)
	return &nft, result.Error
}
