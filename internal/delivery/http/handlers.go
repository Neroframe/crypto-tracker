package http

import (
	"errors"
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

func (h *CryptoHandler) AddCurrency(c *gin.Context) {
	log := h.logger.With("handler", "AddCurrency")

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

func (h *CryptoHandler) RemoveCurrency(c *gin.Context) {

}

func (h *CryptoHandler) GetPrice(c *gin.Context) {

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
