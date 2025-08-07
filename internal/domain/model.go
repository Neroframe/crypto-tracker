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
	ID        uuid.UUID `db:"id"         json:"id"`
	Symbol    string    `db:"symbol"     json:"symbol"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
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
	ID         uuid.UUID `db:"id"            json:"id"`
	CurrencyID uuid.UUID `db:"currency_id"   json:"currency_id"`
	Timestamp  time.Time `db:"timestamp"     json:"timestamp"`
	Price      float64   `db:"price"         json:"price"`
	CreatedAt  time.Time `db:"created_at"    json:"created_at"`
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
