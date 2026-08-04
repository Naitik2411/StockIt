package handler

import (
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
)

type LeagueHandler struct {
	Handler
	authService   *service.AuthService
	leagueService *service.LeagueService
}

func NewLeagueHandler(s *server.Server, authService *service.AuthService, leagueService *service.LeagueService) *LeagueHandler {
	return &LeagueHandler{
		Handler:       NewHandler(s),
		authService:   authService,
		leagueService: leagueService,
	}
}
