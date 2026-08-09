package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/service"
	"github.com/google/uuid"
)

// registerUser registers a fresh user and returns their ID.
func registerUser(t *testing.T, svc *service.Service) uuid.UUID {
	t.Helper()
	user, _, err := svc.RegisterUser(context.Background(), &models.RegisterUser{
		FirstName: "Test",
		LastName:  "User",
		Email:     uniqueEmail(),
		Password:  "supersecret",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user.ID
}

func TestCreateMonitor_Success(t *testing.T) {
	svc := newService(t)
	uid := registerUser(t, svc)

	m, err := svc.CreateMonitor(context.Background(), uid, &models.CreateMonitorRequest{
		Name: "blog", URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("CreateMonitor: %v", err)
	}
	if m.ID == uuid.Nil || m.UserID != uid {
		t.Fatalf("monitor shape wrong: %+v", m)
	}
}

func TestMonitorCRUD_Ownership(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	owner := registerUser(t, svc)
	stranger := registerUser(t, svc)

	created, err := svc.CreateMonitor(ctx, owner, &models.CreateMonitorRequest{
		Name: "blog", URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("CreateMonitor: %v", err)
	}

	// owner can read
	if _, err := svc.GetMonitorByID(ctx, owner, created.ID); err != nil {
		t.Fatalf("owner GetMonitorByID: %v", err)
	}

	// stranger cannot
	if _, err := svc.GetMonitorByID(ctx, stranger, created.ID); !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("stranger GetMonitorByID expected ErrForbidden, got %v", err)
	}

	// stranger cannot update
	newName := "hacked"
	if _, err := svc.UpdateMonitor(ctx, stranger, created.ID, &models.UpdateMonitorRequest{Name: &newName}); !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("stranger UpdateMonitor expected ErrForbidden, got %v", err)
	}

	// stranger cannot delete
	if err := svc.DeleteMonitor(ctx, stranger, created.ID); !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("stranger DeleteMonitor expected ErrForbidden, got %v", err)
	}

	// owner can update
	renamed := "blog-v2"
	updated, err := svc.UpdateMonitor(ctx, owner, created.ID, &models.UpdateMonitorRequest{Name: &renamed})
	if err != nil {
		t.Fatalf("owner UpdateMonitor: %v", err)
	}
	if updated.Name != renamed {
		t.Fatalf("update didn't apply: %+v", updated)
	}

	// GetAllMonitors only returns the owner's monitors
	all, err := svc.GetAllMonitors(ctx, owner)
	if err != nil {
		t.Fatalf("GetAllMonitors: %v", err)
	}
	found := false
	for _, m := range all {
		if m.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("monitor missing from owner's list")
	}

	// owner's list does not include stranger's view
	_, err = svc.GetMonitorByID(ctx, owner, uuid.New())
	if !errors.Is(err, service.ErrMonitorNotFound) {
		t.Fatalf("expected ErrMonitorNotFound, got %v", err)
	}

	// owner can delete
	if err := svc.DeleteMonitor(ctx, owner, created.ID); err != nil {
		t.Fatalf("owner DeleteMonitor: %v", err)
	}
	if _, err := svc.GetMonitorByID(ctx, owner, created.ID); !errors.Is(err, service.ErrMonitorNotFound) {
		t.Fatalf("expected ErrMonitorNotFound after delete, got %v", err)
	}
}

func TestGetMonitorByID_NotFound(t *testing.T) {
	svc := newService(t)
	uid := registerUser(t, svc)
	if _, err := svc.GetMonitorByID(context.Background(), uid, uuid.New()); !errors.Is(err, service.ErrMonitorNotFound) {
		t.Fatalf("expected ErrMonitorNotFound, got %v", err)
	}
}

func TestUpdateMonitor_NotFound(t *testing.T) {
	svc := newService(t)
	uid := registerUser(t, svc)
	if _, err := svc.UpdateMonitor(context.Background(), uid, uuid.New(), &models.UpdateMonitorRequest{}); !errors.Is(err, service.ErrMonitorNotFound) {
		t.Fatalf("expected ErrMonitorNotFound, got %v", err)
	}
}

func TestDeleteMonitor_NotFound(t *testing.T) {
	svc := newService(t)
	uid := registerUser(t, svc)
	if err := svc.DeleteMonitor(context.Background(), uid, uuid.New()); !errors.Is(err, service.ErrMonitorNotFound) {
		t.Fatalf("expected ErrMonitorNotFound, got %v", err)
	}
}
