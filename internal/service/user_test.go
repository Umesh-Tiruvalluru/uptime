package service_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/config"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	sharedPool *pgxpool.Pool
	once       sync.Once
)

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

func newService(t *testing.T) *service.Service {
	t.Helper()
	cfg := &config.Config{JWTSecret: "test-secret-do-not-use-in-prod"}
	return service.NewService(repository.NewRepository(openTestPool(t)), cfg)
}

func uniqueEmail() string {
	return "svc_" + uuid.NewString() + "@example.com"
}

func TestRegisterUser_Success(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	in := &models.RegisterUser{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     uniqueEmail(),
		Password:  "supersecret",
	}
	user, token, err := svc.RegisterUser(ctx, in)
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Fatal("expected user ID")
	}
	if !user.CreatedAt.Before(time.Now().Add(time.Second)) {
		t.Fatalf("CreatedAt seems wrong: %v", user.CreatedAt)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	// Email should be lowercased
	if user.Email != strings.ToLower(in.Email) {
		t.Fatalf("expected lowercased email, got %q", user.Email)
	}
	// Password should be stored as a bcrypt hash, not plaintext
	if strings.HasPrefix(user.Password, "$2") == false {
		t.Fatalf("expected bcrypt hash, got %q", user.Password)
	}
}

func TestRegisterUser_EmailAlreadyTaken(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	email := uniqueEmail()
	_, _, err := svc.RegisterUser(ctx, &models.RegisterUser{
		FirstName: "A", LastName: "B", Email: email, Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("first RegisterUser: %v", err)
	}
	_, _, err = svc.RegisterUser(ctx, &models.RegisterUser{
		FirstName: "C", LastName: "D", Email: email, Password: "supersecret",
	})
	if !errors.Is(err, service.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegisterUser_ShortPassword(t *testing.T) {
	svc := newService(t)
	_, _, err := svc.RegisterUser(context.Background(), &models.RegisterUser{
		FirstName: "A", LastName: "B", Email: uniqueEmail(), Password: "short",
	})
	if !errors.Is(err, service.ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLoginUser_Success(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	email := uniqueEmail()
	pw := "supersecret"
	_, _, err := svc.RegisterUser(ctx, &models.RegisterUser{
		FirstName: "A", LastName: "B", Email: email, Password: pw,
	})
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	token, err := svc.LoginUser(ctx, &models.LoginUser{Email: email, Password: pw})
	if err != nil {
		t.Fatalf("LoginUser: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
}

func TestLoginUser_WrongPassword(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	email := uniqueEmail()
	if _, _, err := svc.RegisterUser(ctx, &models.RegisterUser{
		FirstName: "A", LastName: "B", Email: email, Password: "supersecret",
	}); err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	_, err := svc.LoginUser(ctx, &models.LoginUser{Email: email, Password: "wrong"})
	if !errors.Is(err, service.ErrInvalidCreds) {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestLoginUser_UnknownEmail(t *testing.T) {
	svc := newService(t)
	_, err := svc.LoginUser(context.Background(), &models.LoginUser{
		Email: uniqueEmail(), Password: "anything",
	})
	if !errors.Is(err, service.ErrInvalidCreds) {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestParseToken_RoundTrip(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	user, token, err := svc.RegisterUser(ctx, &models.RegisterUser{
		FirstName: "A", LastName: "B", Email: uniqueEmail(), Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}

	got, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if got != user.ID {
		t.Fatalf("token sub mismatch: %s vs %s", got, user.ID)
	}
}

func TestParseToken_Invalid(t *testing.T) {
	svc := newService(t)
	if _, err := svc.ParseToken("not-a-token"); err == nil {
		t.Fatal("expected error for garbage token")
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	svc := newService(t)
	_, err := svc.GetUserByID(context.Background(), uuid.New())
	if !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
