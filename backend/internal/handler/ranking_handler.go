package handler

import (
	"net/http"
	"strconv"

	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type RankingHandler struct {
	Handler
	eloService *service.ELOService
}

func NewRankingHandler(s *server.Server, eloService *service.ELOService) *RankingHandler {
	return &RankingHandler{
		Handler:    NewHandler(s),
		eloService: eloService,
	}
}

func (h *RankingHandler) GlobalELO(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	result, err := h.eloService.GlobalRankings(c.Request().Context(), page, limit)
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

func (h *RankingHandler) UserELO(c *echo.Context) error {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	detail, err := h.eloService.UserRating(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    detail,
	})
}
