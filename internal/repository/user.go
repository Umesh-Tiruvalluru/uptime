package repository

import (
	"context"
	"errors"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("not found")

func (r *Repository) CreateUser(ctx context.Context, u *models.User) error {
	const q = `
        INSERT INTO users (first_name, last_name, email, password)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at`
	return r.db.QueryRow(ctx, q, u.FirstName, u.LastName, u.Email, u.Password).
		Scan(&u.ID, &u.CreatedAt)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
        SELECT id, first_name, last_name, email, password, created_at
        FROM users WHERE email = $1`
	u := &models.User{}
	err := r.db.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const q = `
        SELECT id, first_name, last_name, email, created_at
        FROM users WHERE id = $1`
	u := &models.User{}
	err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}
