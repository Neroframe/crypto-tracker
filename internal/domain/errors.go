package domain

import "errors"

var (
	ErrInvalidSymbol     = errors.New("invalid cryptocurrency symbol")
	ErrDuplicateCurrency = errors.New("cryptocurrency already exist")
	ErrNotTracked        = errors.New("cryptocurrency not tracked")

	ErrNegativePrice   = errors.New("price must be non-negative")
	ErrTimestampFuture = errors.New("timestamp cannot be in the future")
	ErrDuplicatePrice  = errors.New("price already exists for this timestamp")
	ErrPriceNotFound   = errors.New("price not found in the database")
)
