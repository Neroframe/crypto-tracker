package httpdto

type AddCurrencyRequest struct {
	Symbol string `json:"symbol" validate:"required,uppercase,alphanum,min=1,max=10"`
}

type AddCurrencyResponse struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	CreatedAt string `json:"created_at"` // RFC3339 string
}

type RemoveCurrencyRequest struct {
	Symbol string `json:"symbol" validate:"required,uppercase,alphanum,min=1,max=10"`
}

type RemoveCurrencyResponse struct {
	Message string `json:"message"`
}

type PriceQueryRequest struct {
	Symbol    string `json:"symbol" validate:"required,uppercase,alphanum,min=1,max=10"`
	Timestamp int64  `json:"timestamp" validate:"required,gt=0"`
}

type PriceQueryResponse struct {
	Symbol          string  `json:"symbol"`
	RequestedUnixTs int64   `json:"requested_timestamp"`
	ReturnedUnixTs  int64   `json:"returned_timestamp"`
	Price           float64 `json:"price"`
}
