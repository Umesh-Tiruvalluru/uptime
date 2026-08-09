package repository_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	sharedPool *pgxpool.Pool
	once       sync.Once
)

// openTestPool opens (or returns) a pool against DATABASE_URL_TEST.
// If the env var is unset, the test is skipped.
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	once.Do(func() {
		dsn := os.Getenv("DATABASE_URL_TEST")
		if dsn == "" {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			if err := pool.Ping(ctx); err == nil {
				sharedPool = pool
			}
		}
	})
	if sharedPool == nil {
		t.Skip("DATABASE_URL_TEST not set; skipping integration test")
	}
	return sharedPool
}

func TestUserRepository_CRUD(t *testing.T) {
	pool := openTestPool(t)
	repo := repository.NewRepository(pool)
	ctx := context.Background()

	email := "test_" + uuid.NewString() + "@example.com"
	u := &models.User{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     email,
		Password:  "hashed-pw",
	}

	if err := repo.CreateUser(ctx, u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == uuid.Nil {
		t.Fatal("expected ID to be populated")
	}
	if u.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be populated")
	}

	got, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if got.ID != u.ID || got.Email != email {
		t.Fatalf("GetUserByEmail mismatch: %+v vs %+v", got, u)
	}

	gotByID, err := repo.GetUserByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if gotByID.Email != email {
		t.Fatalf("GetUserByID mismatch: %+v", gotByID)
	}

	if _, err := repo.GetUserByEmail(ctx, "nope_"+uuid.NewString()+"@x"); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMonitorRepository_CRUD(t *testing.T) {
	pool := openTestPool(t)
	repo := repository.NewRepository(pool)
	ctx := context.Background()

	user := newUserFixture(t, repo, ctx)

	m := newMonitorFixture(t, repo, ctx, user.ID, "blog", "https://example.com")
	if m.ID == uuid.Nil {
		t.Fatal("expected monitor ID")
	}

	all, err := repo.GetMonitorsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetMonitorsByUserID: %v", err)
	}
	if len(all) != 1 || all[0].ID != m.ID {
		t.Fatalf("expected 1 monitor for user, got %d", len(all))
	}

	got, err := repo.GetMonitorByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("GetMonitorByID: %v", err)
	}
	if got.URL != m.URL {
		t.Fatalf("URL mismatch: %s vs %s", got.URL, m.URL)
	}

	// partial update
	newName := "blog-renamed"
	updated, err := repo.UpdateMonitor(ctx, m.ID, &models.UpdateMonitorRequest{Name: &newName})
	if err != nil {
		t.Fatalf("UpdateMonitor: %v", err)
	}
	if updated.Name != newName || updated.URL != m.URL {
		t.Fatalf("partial update failed: %+v", updated)
	}
	if !updated.UpdatedAt.After(m.UpdatedAt) {
		t.Fatalf("expected UpdatedAt to advance")
	}

	if err := repo.DeleteMonitor(ctx, m.ID); err != nil {
		t.Fatalf("DeleteMonitor: %v", err)
	}
	if _, err := repo.GetMonitorByID(ctx, m.ID); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMonitorRepository_NotFound(t *testing.T) {
	pool := openTestPool(t)
	repo := repository.NewRepository(pool)
	ctx := context.Background()

	if _, err := repo.GetMonitorByID(ctx, uuid.New()); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := repo.DeleteMonitor(ctx, uuid.New()); err != repository.ErrNotFound {
		t.Fatalf("DeleteMonitor expected ErrNotFound, got %v", err)
	}
	if _, err := repo.UpdateMonitor(ctx, uuid.New(), &models.UpdateMonitorRequest{}); err != repository.ErrNotFound {
		t.Fatalf("UpdateMonitor expected ErrNotFound, got %v", err)
	}
}
