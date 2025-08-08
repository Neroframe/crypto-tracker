package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Neroframe/crypto-tracker/internal/domain"
	"github.com/Neroframe/crypto-tracker/pkg/logger"
)

type CryptoService interface {
	AddCurrency(ctx context.Context, symbol string) (*domain.Currency, error)
	RemoveCurrency(ctx context.Context, symbol string) error
	GetPrice(ctx context.Context, symbol string, at time.Time) (*domain.PriceSnapshot, error)
	FetchAndStorePrices(ctx context.Context) error
}

type cryptoService struct {
	repo domain.CryptoRepository
	api  ExternalPriceAPI
	log  *logger.Logger
}

func NewCryptoService(repo domain.CryptoRepository, api ExternalPriceAPI, log *logger.Logger) CryptoService {
	return &cryptoService{repo: repo, api: api, log: log}
}

func (s *cryptoService) AddCurrency(ctx context.Context, symbol string) (*domain.Currency, error) {
	ok, err := s.api.SymbolExists(symbol)
	if err != nil {
		s.log.Error("checking symbol failed", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to verify symbol: %w", err)
	}
	if !ok {
		return nil, domain.ErrInvalidSymbol
	}

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
