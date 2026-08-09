package repository_test

import (
	"context"
	"testing"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/google/uuid"
)

// newUserFixture inserts a user with a unique email and returns the populated User.
func newUserFixture(t *testing.T, repo *repository.Repository, ctx context.Context) *models.User {
	t.Helper()
	u := &models.User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "fixture_" + uuid.NewString() + "@example.com",
		Password:  "hashed",
	}
	if err := repo.CreateUser(ctx, u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

// newMonitorFixture inserts a monitor owned by the given user.
func newMonitorFixture(t *testing.T, repo *repository.Repository, ctx context.Context, userID uuid.UUID, name, url string) *models.Monitor {
	t.Helper()
	m := &models.Monitor{
		UserID: userID,
		Name:   name,
		URL:    url,
	}
	if err := repo.CreateMonitor(ctx, m); err != nil {
		t.Fatalf("seed monitor: %v", err)
	}
	return m
}
