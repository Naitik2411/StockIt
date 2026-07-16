package router

import (
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/labstack/echo/v5"
)

func registerPortfolioRoutes(v1 *echo.Group, h *handler.Handlers, auth *middleware.AuthMiddleware) {
	protected := v1.Group("", auth.RequireAuth)

	protected.GET("/portfolio", h.Portfolio.Summary)
	protected.GET("/portfolio/positions", h.Portfolio.Positions)
	protected.POST("/portfolio/buy", h.Portfolio.Buy)
	protected.POST("/portfolio/sell", h.Portfolio.Sell)
	protected.GET("/portfolio/transactions", h.Portfolio.Transactions)
}
