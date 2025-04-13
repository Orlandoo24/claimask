package dao

import (
	"claimask/internal/monitor/model/po"

	"gorm.io/gorm"
)

type UTXODao interface {
	GetAddressValidUtxo(address string) (*po.UTXOPO, error)
	UpdateUTXOState(utxo *po.UTXOPO) error
}

type UTXODaoImpl struct {
	db *gorm.DB
}

func NewUTXODao(db *gorm.DB) UTXODao {
	return &UTXODaoImpl{db: db}
}

func (d *UTXODaoImpl) GetAddressValidUtxo(address string) (*po.UTXOPO, error) {
	var utxo po.UTXOPO
	result := d.db.Where("address = ? AND spent = ?", address, false).
		Order("value DESC").First(&utxo)
	if result.Error != nil {
		return nil, result.Error
	}
	return &utxo, nil
}

func (d *UTXODaoImpl) UpdateUTXOState(utxo *po.UTXOPO) error {
	return d.db.Model(utxo).Updates(map[string]interface{}{
		"spent":   utxo.Spent,
		"tx_hash": utxo.TxHash,
		"index":   utxo.Index,
	}).Error
}
