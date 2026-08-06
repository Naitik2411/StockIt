package middleware

import (
	"net/http"

	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const (
	LeagueIDKey   = "league_id"
	LeagueRoleKey = "league_role"
)

type LeagueAdminMiddleware struct {
	server        *server.Server
	authService   *service.AuthService
	leagueService *service.LeagueService
}

func NewLeagueAdminMiddleware(
	s *server.Server,
	authService *service.AuthService,
	leagueService *service.LeagueService,
) *LeagueAdminMiddleware {
	return &LeagueAdminMiddleware{
		server:        s,
		authService:   authService,
		leagueService: leagueService,
	}
}

func (m *LeagueAdminMiddleware) RequireLeagueAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		clerkUserID, ok := c.Get("user_id").(string)
		if !ok || clerkUserID == "" {
			return errorss.NewUnauthorizedError("Unauthorized", false)
		}

		leagueID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid league id")
		}

		user, err := m.authService.CreateOrGetUser(c.Request().Context(), clerkUserID)
		if err != nil {
			return err
		}

		role, err := m.leagueService.GetMemberRole(c.Request().Context(), leagueID, user.ID)
		if err != nil {
			return err
		}

		if role != "admin" {
			m.server.Logger.Warn().
				Str("function", "RequireLeagueAdmin").
				Str("request_id", GetRequestID(c)).
				Str("league_id", leagueID.String()).
				Str("user_id", user.ID.String()).
				Str("role", role).
				Msg("league admin check failed")
			return errorss.NewForbiddenErrorWithCode("not a league admin", "NOT_LEAGUE_ADMIN")
		}

		c.Set(LeagueIDKey, leagueID)
		c.Set(LeagueRoleKey, role)
		return next(c)
	}
}
