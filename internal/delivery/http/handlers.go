package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Neroframe/crypto-tracker/internal/app"
	httpdto "github.com/Neroframe/crypto-tracker/internal/delivery/http/dto"
	"github.com/Neroframe/crypto-tracker/internal/domain"
	"github.com/Neroframe/crypto-tracker/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CryptoHandler struct {
	svc       app.CryptoService
	validator *validator.Validate
	logger    *logger.Logger
}

func NewCryptoHandler(v *validator.Validate, log *logger.Logger, svc app.CryptoService) *CryptoHandler {
	return &CryptoHandler{validator: v, logger: log, svc: svc}
}

// AddCurrency godoc
// @Summary Add a new currency
// @Description Adds a cryptocurrency by symbol to start tracking
// @Tags Currency
// @Accept json
// @Produce json
// @Param input body httpdto.AddCurrencyRequest true "Currency Symbol"
// @Success 201 {object} map[string]httpdto.AddCurrencyResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /currency/add [post]
func (h *CryptoHandler) AddCurrency(c *gin.Context) {
	log := h.logger.With("handler", "AddCurrency")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)

	var req httpdto.AddCurrencyRequest
	if !BindAndValidate(c, h.validator, &req) {
		return
	}

	cur, err := h.svc.AddCurrency(c.Request.Context(), req.Symbol)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSymbol):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrDuplicateCurrency):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Error("service AddCurrency", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	resp := httpdto.AddCurrencyResponse{
		ID:        cur.ID.String(),
		Symbol:    cur.Symbol,
		CreatedAt: cur.CreatedAt.Format(time.RFC3339),
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// RemoveCurrency godoc
// @Summary Remove a tracked currency
// @Description Removes a tracked cryptocurrency by symbol
// @Tags Currency
// @Accept json
// @Produce json
// @Param input body httpdto.RemoveCurrencyRequest true "Currency Symbol"
// @Success 200 {object} map[string]httpdto.RemoveCurrencyResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /currency/remove [post]
func (h *CryptoHandler) RemoveCurrency(c *gin.Context) {
	log := h.logger.With("handler", "RemoveCurrency")
	log.Debug("RemoveCurrency called", "timestamp", time.Now().Format(time.RFC3339Nano))

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)

	var req httpdto.RemoveCurrencyRequest
	if !BindAndValidate(c, h.validator, &req) {
		return
	}

	err := h.svc.RemoveCurrency(c.Request.Context(), req.Symbol)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSymbol):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrNotTracked):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			log.Error("service.RemoveCurrency failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	resp := httpdto.RemoveCurrencyResponse{
		Message: fmt.Sprintf("removed %s", req.Symbol),
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetPrice godoc
// @Summary Get historical price snapshot
// @Description Returns the closest price snapshot at or before the given timestamp
// @Tags Price
// @Accept json
// @Produce json
// @Param input body httpdto.PriceQueryRequest true "Symbol and Unix Timestamp"
// @Success 200 {object} map[string]httpdto.PriceQueryResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /currency/price [post]
func (h *CryptoHandler) GetPrice(c *gin.Context) {
	log := h.logger.With("handler", "GetPrice")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)

	var req httpdto.PriceQueryRequest
	if !BindAndValidate(c, h.validator, &req) {
		return
	}

	// Reject 0 or negative timestamp
	if req.Timestamp <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp must be > 0"})
		return
	}

	ts := time.Unix(req.Timestamp, 0).UTC()
	snap, err := h.svc.GetPrice(c.Request.Context(), req.Symbol, ts)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSymbol):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrPriceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			log.Error("service.GetPrice failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	resp := httpdto.PriceQueryResponse{
		Symbol:          req.Symbol,
		RequestedUnixTs: req.Timestamp,
		ReturnedUnixTs:  snap.Timestamp.Unix(),
		Price:           snap.Price,
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// On error writes the HTTP 400 and returns false
func BindAndValidate(c *gin.Context, validate *validator.Validate, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return false
	}
	if err := validate.Struct(obj); err != nil {
		// pick the first field error
		ve := err.(validator.ValidationErrors)[0]
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ve.Field() + " failed " + ve.Tag(),
		})
		return false
	}
	return true
}
