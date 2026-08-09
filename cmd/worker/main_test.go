package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/google/uuid"
)

func TestPerformHTTPCheck(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantStatus models.MonitorStatus
	}{
		{name: "successful response", statusCode: http.StatusOK, wantStatus: models.StatusUp},
		{name: "redirect response", statusCode: http.StatusNotModified, wantStatus: models.StatusUp},
		{name: "server error", statusCode: http.StatusServiceUnavailable, wantStatus: models.StatusDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			result := performHTTPCheck(context.Background(), server.Client(), uuid.New(), server.URL)
			if result.Status != tt.wantStatus || result.StatusCode != tt.statusCode {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
}

func TestPerformHTTPCheckTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 10 * time.Millisecond}
	result := performHTTPCheck(context.Background(), client, uuid.New(), server.URL)
	if result.Status != models.StatusDown || result.ErrorMessage == "" {
		t.Fatalf("expected timed-out down result, got %+v", result)
	}
}
