package dao

import (
	"astro-orderx/internal/claimask/model/po"

	"github.com/jinzhu/gorm"
)

// OrderDAO 订单DAO接口
type OrderDAO interface {
	CreateOrder(order *po.Order) error
}

// OrderDAOImpl 订单DAO实现
type OrderDAOImpl struct {
	DB *gorm.DB
}

// NewOrderDAO 创建新的订单DAO实例
func NewOrderDAO(db *gorm.DB) OrderDAO {
	return &OrderDAOImpl{DB: db}
}

// CreateOrder 在数据库中创建订单
func (dao *OrderDAOImpl) CreateOrder(order *po.Order) error {
	return dao.DB.Table("order_id").Create(order).Error
}
