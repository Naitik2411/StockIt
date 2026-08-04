package handler

import (
	"net/http"

	"github.com/Naitik2411/stockit/internal/middleware"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type LeagueHandler struct {
	Handler
	authService   *service.AuthService
	leagueService *service.LeagueService
}

type joinLeagueRequest struct {
	InviteCode string `json:"inviteCode"`
}

func NewLeagueHandler(s *server.Server, authService *service.AuthService, leagueService *service.LeagueService) *LeagueHandler {
	return &LeagueHandler{
		Handler:       NewHandler(s),
		authService:   authService,
		leagueService: leagueService,
	}
}

type createLeagueRequest struct {
	Name       string `json:"name"`
	MaxMembers int    `json:"maxMembers"`
	SeasonDays int    `json:"seasonDays"`
}

func (h *LeagueHandler) resolveUserID(c *echo.Context) (uuid.UUID, error) {
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

func (h *LeagueHandler) Create(c *echo.Context) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return err
	}

	var req createLeagueRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	result, err := h.leagueService.Create(c.Request().Context(), userID, service.CreateLeagueInput{
		Name:       req.Name,
		MaxMembers: req.MaxMembers,
		SeasonDays: req.SeasonDays,
	})

	if err != nil {
		middleware.GetLogger(c).Error().Err(err).Msg("create league failed")
		return err
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

func (h *LeagueHandler) Join(c *echo.Context) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	var req joinLeagueRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	league, err := h.leagueService.Join(c.Request().Context(), userID, req.InviteCode)
	if err != nil {
		middleware.GetLogger(c).Error().Err(err).Msg("unable to join league")
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    league,
	})
}

func (h *LeagueHandler) ListMine(c *echo.Context) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return err
	}
	leagues, err := h.leagueService.ListMine(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    leagues,
	})
}
func (h *LeagueHandler) GetByID(c *echo.Context) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return err
	}
	leagueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid league id")
	}
	detail, err := h.leagueService.GetByID(c.Request().Context(), userID, leagueID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    detail,
	})
}
func (h *LeagueHandler) GetMembers(c *echo.Context) error {
	userID, err := h.resolveUserID(c)
	if err != nil {
		return err
	}
	leagueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid league id")
	}
	members, err := h.leagueService.GetMembers(c.Request().Context(), userID, leagueID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    members,
	})
}
