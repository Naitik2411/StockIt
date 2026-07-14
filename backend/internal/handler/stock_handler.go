package handler

import (
	"net/http"

	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/labstack/echo/v5"
)

type StockHandler struct {
	Handler
	stockService *service.StockService
}

func NewStockHandler(s *server.Server, stockService *service.StockService) *StockHandler {
	return &StockHandler{
		Handler:      NewHandler(s),
		stockService: stockService,
	}
}

func (h *StockHandler) GetTicker(c *echo.Context) error {
	ticker := c.Param("ticker")

	price, err := h.stockService.GetTicker(c.Request().Context(), ticker)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    price,
	})
}
