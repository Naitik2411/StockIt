package handler

import (
	"net/http"

	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	Handler
	authService *service.AuthService
}

func NewAuthHandler(s *server.Server, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		Handler:     NewHandler(s),
		authService: authService,
	}
}

func (h *AuthHandler) Me(c *echo.Context) error {
	// middleware sets this to Clerk's subject ID (e.g. user_2abc...)
	clerkUserID, ok := c.Get("user_id").(string)
	if !ok || clerkUserID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing clerk user id")
	}
	user, err := h.authService.CreateOrGetUser(c.Request().Context(), clerkUserID)
	if err != nil {
		middleware.GetLogger(c).Error().Err(err).Msg("failed to get or create user")
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    user,
	})
}
