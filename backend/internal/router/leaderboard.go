package router

import (
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/labstack/echo/v5"
)

func registerLeaderboardRoutes(v1 *echo.Group, h *handler.Handlers, auth *middleware.AuthMiddleware) {
	// Public
	v1.GET("/leaderboard", h.Leaderboard.Global)
	// Protected
	protected := v1.Group("", auth.RequireAuth)
	protected.GET("/leaderboard/me", h.Leaderboard.Me)
}
