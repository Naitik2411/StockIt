package handler

import (
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	Auth    *AuthHandler
	Stock   *StockHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		Auth:    NewAuthHandler(s, services.Auth),
		Stock:   NewStockHandler(s, services.Stock),
	}
}
