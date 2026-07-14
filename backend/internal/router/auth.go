package router

import (
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/labstack/echo/v5"
)

func registerAuthRoutes(v1 *echo.Group, h *handler.Handlers, auth *middleware.AuthMiddleware) {
	protected := v1.Group("", auth.RequireAuth)
	protected.GET("/auth/me", h.Auth.Me)
}
