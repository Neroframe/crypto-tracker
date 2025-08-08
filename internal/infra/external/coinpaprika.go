package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Neroframe/crypto-tracker/internal/app"
	"github.com/Neroframe/crypto-tracker/pkg/logger"
	"golang.org/x/time/rate"
)

type CoinPaprikaClient struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	baseURL    string
	idMap      map[string]string // e.g. "BTC" -> "btc-bitcoin"
	log        *logger.Logger
}

// ApiKey empty for free tier usage
func NewCoinPaprikaClient(
	httpClient *http.Client,
	baseURL string,
	rateLimit float64,
	log *logger.Logger,
) (*CoinPaprikaClient, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	c := &CoinPaprikaClient{
		httpClient: httpClient,
		limiter:    rate.NewLimiter(rate.Limit(rateLimit), 1),
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		idMap:      make(map[string]string),
		log:        log,
	}
	if err := c.populateIDMap(context.Background()); err != nil {
		return nil, fmt.Errorf("populate CoinPaprika ID map: %w", err)
	}
	return c, nil
}

func (c *CoinPaprikaClient) SymbolExists(symbol string) (bool, error) {
	_, ok := c.idMap[strings.ToUpper(symbol)]
	return ok, nil
}

func (c *CoinPaprikaClient) FetchPrice(ctx context.Context, symbol string) (*app.PricePoint, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%w: rate limit: %v", ErrExternalAPI, err)
	}

	id, ok := c.idMap[strings.ToUpper(symbol)]
	if !ok {
		return nil, ErrUnknownSymbol
	}

	url := fmt.Sprintf("%s/v1/tickers/%s", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: new request: %v", ErrExternalAPI, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: request failed: %v", ErrExternalAPI, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrExternalAPI, resp.StatusCode)
	}

	var payload struct {
		Quotes map[string]struct {
			Price float64 `json:"price"`
		} `json:"quotes"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: decode JSON: %v", ErrExternalAPI, err)
	}

	usdQ, ok := payload.Quotes["USD"]
	if !ok {
		return nil, fmt.Errorf("%w: no USD quote for %s", ErrExternalAPI, symbol)
	}

	// return PricePoint
	return &app.PricePoint{
		Symbol:    symbol,
		Price:     usdQ.Price,
		Timestamp: time.Now().UTC().Unix(),
	}, nil
}

// Calls the CoinPaprika /v1/coins endpoint and builds
// a map from uppercase symbol "BTC" to CoinPaprika ID "btc-bitcoin"
func (c *CoinPaprikaClient) populateIDMap(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/coins", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("populateIDMap: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("populateIDMap: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("populateIDMap: bad status %d", resp.StatusCode)
	}

	var coins []struct {
		ID       string `json:"id"`
		Symbol   string `json:"symbol"`
		Type     string `json:"type"`
		IsActive bool   `json:"is_active"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&coins); err != nil {
		return fmt.Errorf("populateIDMap: decode JSON: %w", err)
	}

	// Fill the map only with active coins
	for _, coin := range coins {
		if coin.Type == "coin" && coin.IsActive {
			sym := strings.ToUpper(coin.Symbol)
			// Only set if not already present
			if _, exists := c.idMap[sym]; !exists {
				c.idMap[sym] = coin.ID
			}
		}
	}

	c.log.Debug("populateIDMap sample", "BTC", c.idMap["BTC"])
	c.log.Info("populateIDMap: loaded symbol map", "count", len(c.idMap))
	return nil
}
