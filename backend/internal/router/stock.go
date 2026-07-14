package router

import (
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/labstack/echo/v5"
)

func registerStockRoutes(v1 *echo.Group, h *handler.Handlers) {
	v1.GET("/stocks/:ticker", h.Stock.GetTicker)
}
