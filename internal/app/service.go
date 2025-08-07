package app

import (
	"context"
	"time"

	"github.com/Neroframe/crypto-tracker/internal/domain"
)

type CryptoService interface {
	AddCurrency(ctx context.Context, symbol string) (*domain.Currency, error)
	RemoveCurrency(ctx context.Context, symbol string) error
	GetPrice(ctx context.Context, symbol string, at time.Time) (*domain.PriceSnapshot, error)
}

type cryptoService struct {
	repo domain.CryptoRepository
}

func NewCryptoService(repo domain.CryptoRepository) CryptoService {
	return &cryptoService{repo: repo}
}

func (s *cryptoService) AddCurrency(ctx context.Context, symbol string) (*domain.Currency, error) {
	cur, err := domain.NewCurrency(symbol)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AddCurrency(ctx, cur); err != nil {
		return nil, err
	}
	return cur, nil
}

func (s *cryptoService) RemoveCurrency(ctx context.Context, symbol string) error {
	return s.repo.RemoveCurrency(ctx, symbol)
}

func (s *cryptoService) GetPrice(ctx context.Context, symbol string, at time.Time) (*domain.PriceSnapshot, error) {
	snap, err := s.repo.GetPriceSnapshot(ctx, symbol, at)
	if err != nil {
		return nil, err
	}
	return snap, nil
}
