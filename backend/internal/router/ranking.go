package router

import (
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/labstack/echo/v5"
)

func registerRankingRoutes(v1 *echo.Group, h *handler.Handlers) {
	v1.GET("/rankings/elo", h.Ranking.GlobalELO)
	v1.GET("/rankings/elo/:userID", h.Ranking.UserELO)
}
