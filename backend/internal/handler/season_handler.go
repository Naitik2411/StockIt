package handler

import (
	"net/http"
	"strconv"

	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type SeasonHandler struct {
	Handler
	authService   *service.AuthService
	seasonService *service.SeasonService
}

func NewSeasonHandler(
	s *server.Server,
	authService *service.AuthService,
	seasonService *service.SeasonService,
) *SeasonHandler {
	return &SeasonHandler{
		Handler:       NewHandler(s),
		authService:   authService,
		seasonService: seasonService,
	}
}

type startNextSeasonRequest struct {
	SeasonDays int `json:"seasonDays"`
}

func (h *SeasonHandler) resolveUserID(c *echo.Context) (uuid.UUID, error) {
	clerkUserID, ok := c.Get("user_id").(string)
	if !ok || clerkUserID == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "missing clerk user id")
	}

	user, err := h.authService.CreateOrGetUser(c.Request().Context(), clerkUserID)
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func (h *SeasonHandler) GetCurrent(c *echo.Context) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return err
	}

	leagueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid league id")
	}

	current, err := h.seasonService.GetCurrent(c.Request().Context(), userID, leagueID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    current,
	})
}

func (h *SeasonHandler) GetHistory(c *echo.Context) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return err
	}

	leagueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid league id")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	history, err := h.seasonService.GetHistory(c.Request().Context(), userID, leagueID, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    history.Seasons,
		"meta": map[string]interface{}{
			"page":  history.Page,
			"limit": history.Limit,
			"total": history.Total,
		},
	})
}

// StartNext is guarded by RequireLeagueAdmin, which sets the league id in context.
func (h *SeasonHandler) StartNext(c *echo.Context) error {
	leagueID, ok := c.Get(middleware.LeagueIDKey).(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid league id")
	}

	var req startNextSeasonRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	season, err := h.seasonService.StartNext(c.Request().Context(), leagueID, req.SeasonDays)
	if err != nil {
		middleware.GetLogger(c).Error().Err(err).Msg("start next season failed")
		return err
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    season,
	})
}
