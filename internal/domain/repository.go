package domain

import (
	"context"
	"time"
)

type CryptoRepository interface {
	AddCurrency(ctx context.Context, c *Currency) error
	RemoveCurrency(ctx context.Context, symbol string) error
	GetPriceSnapshot(ctx context.Context, symbol string, ts time.Time) (*PriceSnapshot, error)
	SavePriceSnapshot(ctx context.Context, snap *PriceSnapshot) error
	ListCurrencies(ctx context.Context) ([]*Currency, error)
}
