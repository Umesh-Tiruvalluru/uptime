package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken      = errors.New("email already in use")
	ErrInvalidCreds    = errors.New("invalid email or password")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrUserNotFound    = errors.New("user not found")
)

func (s *Service) RegisterUser(ctx context.Context, in *models.RegisterUser) (*models.User, string, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if len(in.Password) < 8 {
		return nil, "", ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	u := &models.User{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		Password:  string(hash),
	}

	if err := s.repo.CreateUser(ctx, u); err != nil {
		// 23505 = unique_violation
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "users_email_key") {
			return nil, "", ErrEmailTaken
		}
		return nil, "", err
	}

	token, err := s.signToken(u.ID)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

func (s *Service) LoginUser(ctx context.Context, in *models.LoginUser) (string, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	u, err := s.repo.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCreds
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return "", ErrInvalidCreds
	}

	return s.signToken(u.ID)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *Service) signToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) ParseToken(tokenStr string) (uuid.UUID, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !tok.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}
	sub, _ := claims["sub"].(string)
	return uuid.Parse(sub)
}
