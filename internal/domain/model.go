package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// allow 1–10 uppercase alphanumeric chars
var symbolRegex = regexp.MustCompile(`^[A-Z0-9]{1,10}$`)

type Currency struct {
	ID        uuid.UUID
	Symbol    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCurrency(raw string) (*Currency, error) {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if !symbolRegex.MatchString(s) {
		return nil, ErrInvalidSymbol
	}
	now := time.Now().UTC()
	return &Currency{
		ID:        uuid.New(),
		Symbol:    s,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type PriceSnapshot struct {
	ID         uuid.UUID
	CurrencyID uuid.UUID
	Timestamp  time.Time
	Price      float64
	CreatedAt  time.Time
}

func NewPriceSnapshot(curID uuid.UUID, ts time.Time, val float64) (*PriceSnapshot, error) {
	if val < 0 {
		return nil, ErrNegativePrice
	}
	now := time.Now().UTC()
	if ts.After(now) {
		return nil, ErrTimestampFuture
	}
	return &PriceSnapshot{
		ID:         uuid.New(),
		CurrencyID: curID,
		Timestamp:  ts.UTC(),
		Price:      val,
		CreatedAt:  now,
	}, nil
}
