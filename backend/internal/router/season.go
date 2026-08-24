package router

import (
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/labstack/echo/v5"
)

func registerSeasonRoutes(
	v1 *echo.Group,
	h *handler.Handlers,
	auth *middleware.AuthMiddleware,
	leagueAdmin *middleware.LeagueAdminMiddleware,
) {
	protected := v1.Group("", auth.RequireAuth)

	protected.GET("/leagues/:id/seasons", h.Season.GetHistory)
	protected.GET("/leagues/:id/seasons/current", h.Season.GetCurrent)
	protected.POST("/leagues/:id/seasons/next", h.Season.StartNext, leagueAdmin.RequireLeagueAdmin)
}
