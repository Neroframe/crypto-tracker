package postgres

import (
	"time"

	"github.com/google/uuid"
)

type CurrencyModel struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	Symbol    string    `gorm:"column:symbol;type:varchar(10);unique;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (CurrencyModel) TableName() string {
	return "currencies"
}

type PriceSnapshotModel struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CurrencyID uuid.UUID `gorm:"column:currency_id;type:uuid;not null"`
	Timestamp  int64     `gorm:"column:timestamp;not null"`
	Price      float64   `gorm:"column:price;type:numeric(20,10);not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PriceSnapshotModel) TableName() string {
	return "currency_prices"
}
