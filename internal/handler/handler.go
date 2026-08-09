package handler

import (
	"github.com/Umesh-Tiruvalluru/monitoring/internal/config"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/service"
)

type Handler struct {
	svc *service.Service
	cfg *config.Config
}

func NewHandler(svc *service.Service, cfg *config.Config) *Handler {
	return &Handler{
		svc: svc,
		cfg: cfg,
	}
}