package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Neroframe/crypto-tracker/internal/domain"
)

// Holds the fetched data
type PricePoint struct {
	Symbol    string  // e.g. "BTC"
	Price     float64 // USD
	Timestamp int64   // Unix seconds when fetched
}

type ExternalPriceAPI interface {
	FetchPrice(ctx context.Context, symbol string) (*PricePoint, error)
	SymbolExists(symbol string) (bool, error)
}

func (s *cryptoService) FetchAndStorePrices(ctx context.Context) error {
	const pageSize = 100
	offset := 0

	for {
		currs, err := s.repo.ListCurrencies(ctx, pageSize, offset)
		if err != nil {
			return fmt.Errorf("FetchAndStorePrices: list currencies: %w", err)
		}
		if len(currs) == 0 {
			break // no more rows
		}

		for _, cur := range currs {
			pt, err := s.api.FetchPrice(ctx, cur.Symbol)
			if err != nil {
				s.log.Error("fetch price failed", "symbol", cur.Symbol, "error", err)
				continue
			}

			ts := time.Unix(pt.Timestamp, 0).UTC()
			snap, err := domain.NewPriceSnapshot(cur.ID, ts, pt.Price)
			if err != nil {
				s.log.Error("invalid price snapshot", "symbol", cur.Symbol, "error", err)
				continue
			}

			if err := s.repo.SavePriceSnapshot(ctx, snap); err != nil {
				s.log.Error("save snapshot failed", "symbol", cur.Symbol, "timestamp", snap.Timestamp, "error", err)
				continue
			}

			s.log.Debug("saved snapshot", "symbol", cur.Symbol, "price", snap.Price, "timestamp", snap.Timestamp)
		}

		offset += pageSize
	}

	return nil
}
