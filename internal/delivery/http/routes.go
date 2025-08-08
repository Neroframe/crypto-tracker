package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(h *CryptoHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	currency := r.Group("/currency")
	{
		currency.POST("/add", h.AddCurrency)
		currency.POST("/remove", h.RemoveCurrency)
		currency.POST("/price", h.GetPrice)
	}

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	return r
}
