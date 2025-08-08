package external

import "errors"

var (
	ErrUnknownSymbol     = errors.New("unknown symbol")
	ErrExternalRateLimit = errors.New("rate limit exceeded")
	ErrExternalAPI       = errors.New("external API error")
	ErrPriceNotFound     = errors.New("price not found")
)
