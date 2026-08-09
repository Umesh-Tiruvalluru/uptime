package service

import (
	"context"
	"errors"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrMonitorNotFound = errors.New("monitor not found")
	ErrForbidden       = errors.New("you do not own this monitor")
)

func (s *Service) CreateMonitor(ctx context.Context, userID uuid.UUID, in *models.CreateMonitorRequest) (*models.Monitor, error) {
	m := &models.Monitor{
		UserID:       userID,
		Name:         in.Name,
		URL:          in.URL,
		IntervalSecs: in.IntervalSecs,
	}
	if err := s.repo.CreateMonitor(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) GetAllMonitors(ctx context.Context, userID uuid.UUID) ([]models.Monitor, error) {
	return s.repo.GetMonitorsByUserID(ctx, userID)
}

func (s *Service) GetPublicMonitors(ctx context.Context) ([]models.PublicMonitor, error) {
	return s.repo.GetPublicMonitors(ctx)
}

func (s *Service) GetMonitorByID(ctx context.Context, userID, id uuid.UUID) (*models.Monitor, error) {
	m, err := s.repo.GetMonitorByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrMonitorNotFound
		}
		return nil, err
	}
	if m.UserID != userID {
		return nil, ErrForbidden
	}
	return m, nil
}

func (s *Service) UpdateMonitor(ctx context.Context, userID, id uuid.UUID, in *models.UpdateMonitorRequest) (*models.Monitor, error) {
	existing, err := s.repo.GetMonitorByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrMonitorNotFound
		}
		return nil, err
	}
	if existing.UserID != userID {
		return nil, ErrForbidden
	}

	return s.repo.UpdateMonitor(ctx, id, in)
}

func (s *Service) DeleteMonitor(ctx context.Context, userID, id uuid.UUID) error {
	existing, err := s.repo.GetMonitorByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrMonitorNotFound
		}
		return err
	}
	if existing.UserID != userID {
		return ErrForbidden
	}
	return s.repo.DeleteMonitor(ctx, id)
}

func (s *Service) GetCheckHistory(ctx context.Context, userID, id uuid.UUID, limit int) ([]models.CheckResult, error) {
	m, err := s.repo.GetMonitorByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrMonitorNotFound
		}
		return nil, err
	}
	if userID != uuid.Nil && m.UserID != userID {
		return nil, ErrForbidden
	}
	return s.repo.GetCheckHistory(ctx, id, limit)
}

