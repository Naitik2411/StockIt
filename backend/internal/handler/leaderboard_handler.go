package handler

import (
	"net/http"
	"strconv"

	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/labstack/echo/v5"
)

type LeaderboardHandler struct {
	Handler
	authService        *service.AuthService
	leaderboardService *service.LeaderboardService
}

func NewLeaderboardHandler(
	s *server.Server,
	authService *service.AuthService,
	leaderboardService *service.LeaderboardService,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		Handler:            NewHandler(s),
		authService:        authService,
		leaderboardService: leaderboardService,
	}
}

func (h *LeaderboardHandler) Global(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	result, err := h.leaderboardService.Global(c.Request().Context(), page, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result.Entries,
		"meta": map[string]interface{}{
			"page":  result.Page,
			"limit": result.Limit,
			"total": result.Total,
		},
	})
}

func (h *LeaderboardHandler) Me(c *echo.Context) error {
	clerkUserID, ok := c.Get("user_id").(string)
	if !ok || clerkUserID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing clerk user id")
	}
	user, err := h.authService.CreateOrGetUser(c.Request().Context(), clerkUserID)
	if err != nil {
		return err
	}
	result, err := h.leaderboardService.MyRank(c.Request().Context(), user.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}
