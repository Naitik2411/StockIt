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

type PortfolioHandler struct {
	Handler
	authService      *service.AuthService
	portfolioService *service.PortfolioService
}

func NewPortfolioHandler(
	s *server.Server,
	authService *service.AuthService,
	portfolioService *service.PortfolioService,
) *PortfolioHandler {
	return &PortfolioHandler{
		Handler:          NewHandler(s),
		authService:      authService,
		portfolioService: portfolioService,
	}
}

type tradeRequest struct {
	Ticker string `json:"ticker"`
	Shares string `json:"shares"`
}

func (h *PortfolioHandler) resolveUserID(c *echo.Context) (string, error) {
	clerkUserID, ok := c.Get("user_id").(string)
	if !ok || clerkUserID == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "missing clerk user id")
	}

	user, err := h.authService.CreateOrGetUser(c.Request().Context(), clerkUserID)
	if err != nil {
		return "", err
	}
	return user.ID.String(), nil
}

func (h *PortfolioHandler) Summary(c *echo.Context) error {
	userIDStr, err := h.resolveUserID(c)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}
	summary, err := h.portfolioService.Summary(c.Request().Context(), userID)
	if err != nil {
		middleware.GetLogger(c).Error().Err(err).Msg("portfolio summary failed")
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

func (h *PortfolioHandler) Positions(c *echo.Context) error {
	userIDStr, err := h.resolveUserID(c)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	positions, err := h.portfolioService.Positions(c.Request().Context(), userId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    positions,
	})
}

func (h *PortfolioHandler) Buy(c *echo.Context) error {
	userIDStr, err := h.resolveUserID(c)
	if err != nil {
		return err
	}

	user, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	var req tradeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	if err := h.portfolioService.Buy(c.Request().Context(), user, req.Ticker, req.Shares); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "buy order completed"},
	})
}

func (h *PortfolioHandler) Sell(c *echo.Context) error {
	userIDStr, err := h.resolveUserID(c)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}

	var req tradeRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.portfolioService.Sell(c.Request().Context(), userID, req.Ticker, req.Shares); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "sell order completed"},
	})
}

func (h *PortfolioHandler) Transactions(c *echo.Context) error {
	userIDStr, err := h.resolveUserID(c)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	txns, err := h.portfolioService.Transactions(c.Request().Context(), userID, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    txns,
	})
}
