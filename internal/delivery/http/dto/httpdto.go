package httpdto

type AddCurrencyRequest struct {
	Symbol string `json:"symbol" example:"BTC" validate:"required,uppercase,alphanum,min=1,max=10"`
}

type AddCurrencyResponse struct {
	ID        string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Symbol    string `json:"symbol" example:"BTC"`
	CreatedAt string `json:"created_at" example:"2025-08-08T18:00:00Z"`
}

type RemoveCurrencyRequest struct {
	Symbol string `json:"symbol" example:"BTC" validate:"required,uppercase,alphanum,min=1,max=10"`
}

type RemoveCurrencyResponse struct {
	Message string `json:"message" example:"removed BTC"`
}

type PriceQueryRequest struct {
	Symbol    string `json:"symbol" example:"BTC" validate:"required,uppercase,alphanum,min=1,max=10"`
	Timestamp int64  `json:"timestamp" example:"1723123200" validate:"required,gt=0"`
}

type PriceQueryResponse struct {
	Symbol          string  `json:"symbol" example:"BTC"`
	RequestedUnixTs int64   `json:"requested_timestamp" example:"1723123200"`
	ReturnedUnixTs  int64   `json:"returned_timestamp" example:"1723123199"`
	Price           float64 `json:"price" example:"29753.55"`
}
