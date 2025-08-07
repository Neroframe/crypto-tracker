package domain

import "errors"

var (
	ErrInvalidSymbol     = errors.New("invalid cryptocurrency symbol")
	ErrDuplicateCurrency = errors.New("cryptocurrency already exist")

	ErrNegativePrice   = errors.New("price must be non-negative")
	ErrTimestampFuture = errors.New("timestamp cannot be in the future")
	ErrDuplicatePrice  = errors.New("price already exists for this timestamp")
)
