package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Neroframe/crypto-tracker/internal/domain"
	"github.com/Neroframe/crypto-tracker/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type GormRepo struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewGormRepo(db *gorm.DB, log *logger.Logger) *GormRepo {
	return &GormRepo{db: db, logger: log}
}

func (r *GormRepo) AddCurrency(ctx context.Context, c *domain.Currency) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Rely on autoCreateTime/autoUpdateTime tags
		model := CurrencyModel{
			ID:     c.ID,
			Symbol: c.Symbol,
		}
		if err := tx.Create(&model).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrDuplicateCurrency
			}
			return err
		}
		return nil
	})

	if err != nil {
		r.logger.Error("repo.AddCurrency failed", "symbol", c.Symbol, "error", err)
		return fmt.Errorf("AddCurrency: %w", err)
	}
	return nil
}

func (r *GormRepo) RemoveCurrency(ctx context.Context, symbol string) error {
	res := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Delete(&CurrencyModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotTracked
	}
	return nil
}

// Finds the nearest price with closest to `ts` (before or after)
func (r *GormRepo) GetPriceSnapshot(ctx context.Context, symbol string, ts time.Time) (*domain.PriceSnapshot, error) {
	var cm CurrencyModel
	if err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		First(&cm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotTracked
		}
		return nil, fmt.Errorf("gorm GetPriceSnapshot (currency lookup): %w", err)
	}

	tsUnix := ts.Unix()

	// Try exact match
	var exact PriceSnapshotModel
	if err := r.db.WithContext(ctx).
		Where("currency_id = ? AND timestamp = ?", cm.ID, tsUnix).
		First(&exact).Error; err == nil {
		return &domain.PriceSnapshot{
			ID:         exact.ID,
			CurrencyID: exact.CurrencyID,
			Timestamp:  time.Unix(exact.Timestamp, 0).UTC(),
			Price:      exact.Price,
			CreatedAt:  exact.CreatedAt,
		}, nil
	}

	// Fallback to nearest older, newer
	var older, newer PriceSnapshotModel
	errOlder := r.db.WithContext(ctx).
		Where("currency_id = ? AND timestamp < ?", cm.ID, tsUnix).
		Order("timestamp DESC").Limit(1).First(&older).Error

	errNewer := r.db.WithContext(ctx).
		Where("currency_id = ? AND timestamp > ?", cm.ID, tsUnix).
		Order("timestamp ASC").Limit(1).First(&newer).Error

	if errors.Is(errOlder, gorm.ErrRecordNotFound) && errors.Is(errNewer, gorm.ErrRecordNotFound) {
		return nil, domain.ErrPriceNotFound
	}

	var chosen PriceSnapshotModel
	switch {
	case errOlder == nil && errNewer == nil:
		if (tsUnix - older.Timestamp) <= (newer.Timestamp - tsUnix) {
			chosen = older
		} else {
			chosen = newer
		}
	case errOlder == nil:
		chosen = older
	case errNewer == nil:
		chosen = newer
	}

	return &domain.PriceSnapshot{
		ID:         chosen.ID,
		CurrencyID: chosen.CurrencyID,
		Timestamp:  time.Unix(chosen.Timestamp, 0).UTC(),
		Price:      chosen.Price,
		CreatedAt:  chosen.CreatedAt,
	}, nil
}

func (r *GormRepo) SavePriceSnapshot(ctx context.Context, snap *domain.PriceSnapshot) error {
	model := PriceSnapshotModel{
		ID:         snap.ID,
		CurrencyID: snap.CurrencyID,
		Timestamp:  snap.Timestamp.Unix(),
		Price:      snap.Price,
		CreatedAt:  snap.CreatedAt,
	}

	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicatePrice
		}
		return fmt.Errorf("gorm SavePriceSnapshot: %w", err)
	}

	return nil
}

func (r *GormRepo) ListPriceSnapshots(ctx context.Context, currencyID uuid.UUID, start, end time.Time) ([]*domain.PriceSnapshot, error) {
	var rows []PriceSnapshotModel

	err := r.db.WithContext(ctx).
		Where("currency_id = ? AND timestamp BETWEEN ? AND ?", currencyID, start.Unix(), end.Unix()).
		Order("timestamp ASC").
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("gorm ListPriceSnapshots: %w", err)
	}

	snaps := make([]*domain.PriceSnapshot, len(rows))
	for i, pm := range rows {
		snaps[i] = &domain.PriceSnapshot{
			ID:         pm.ID,
			CurrencyID: pm.CurrencyID,
			Timestamp:  time.Unix(pm.Timestamp, 0).UTC(),
			Price:      pm.Price,
			CreatedAt:  pm.CreatedAt,
		}
	}
	return snaps, nil
}

func (r *GormRepo) ListCurrencies(ctx context.Context, limit, offset int) ([]*domain.Currency, error) {
	if limit <= 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination params: limit=%d offset=%d", limit, offset)
	}

	var rows []CurrencyModel
	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("ListCurrencies: %w", err)
	}

	currs := make([]*domain.Currency, len(rows))
	for i, cm := range rows {
		currs[i] = &domain.Currency{
			ID:        cm.ID,
			Symbol:    cm.Symbol,
			CreatedAt: cm.CreatedAt,
			UpdatedAt: cm.UpdatedAt,
		}
	}

	return currs, nil
}
