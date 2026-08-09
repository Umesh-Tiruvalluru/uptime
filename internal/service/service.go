package service

import (
	"github.com/Umesh-Tiruvalluru/monitoring/internal/config"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
)

type Service struct {
	repo *repository.Repository
	cfg  *config.Config

}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		repo: repo,
		cfg:  cfg,
	}
}