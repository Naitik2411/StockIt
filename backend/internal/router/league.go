package router

import (
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/labstack/echo/v5"
)

func registerLeagueRoutes(v1 *echo.Group, h *handler.Handlers, auth *middleware.AuthMiddleware) {
	protected := v1.Group("", auth.RequireAuth)
	protected.POST("/leagues", h.League.Create)
	protected.GET("/leagues", h.League.ListMine)
	protected.POST("/leagues/join", h.League.Join)
	protected.GET("/leagues/:id", h.League.GetByID)
	protected.GET("/leagues/:id/members", h.League.GetMembers)
}
