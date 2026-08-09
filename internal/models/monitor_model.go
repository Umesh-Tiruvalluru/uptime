package models

import (
	"time"

	"github.com/google/uuid"
)

type MonitorStatus string

const (
	StatusPending MonitorStatus = "pending"
	StatusUp      MonitorStatus = "up"
	StatusDown    MonitorStatus = "down"
)

type Monitor struct {
	ID             uuid.UUID     `json:"id"`
	UserID         uuid.UUID     `json:"userId"`
	Name           string        `json:"name"`
	URL            string        `json:"url"`
	IntervalSecs   int           `json:"intervalSeconds"`
	NextCheckAt    time.Time     `json:"nextCheckAt"`
	LastCheckedAt  *time.Time    `json:"lastCheckedAt,omitempty"`
	LastStatus     MonitorStatus `json:"lastStatus"`
	LastResponseMs *int64        `json:"lastResponseMs,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// PublicMonitor is the limited monitor shape exposed by the unauthenticated
// public status endpoint.
type PublicMonitor struct {
	ID             uuid.UUID     `json:"id"`
	Name           string        `json:"name"`
	URL            string        `json:"url"`
	LastStatus     MonitorStatus `json:"lastStatus"`
	LastCheckedAt  *time.Time    `json:"lastCheckedAt,omitempty"`
	LastResponseMs *int64        `json:"lastResponseMs,omitempty"`
}

type CreateMonitorRequest struct {
	Name         string `json:"name" binding:"required"`
	URL          string `json:"url" binding:"required,url"`
	IntervalSecs int    `json:"intervalSeconds" binding:"required,min=10,max=86400"`
}

type UpdateMonitorRequest struct {
	Name         *string `json:"name"`
	URL          *string `json:"url" binding:"omitempty,url"`
	IntervalSecs *int    `json:"intervalSeconds" binding:"omitempty,min=10,max=86400"`
	Paused       *bool   `json:"paused"`
}

// CheckResult is what the worker publishes to NATS after a check completes.
type CheckResult struct {
	MonitorID    uuid.UUID     `json:"monitorId"`
	Status       MonitorStatus `json:"status"`
	StatusCode   int           `json:"statusCode"`
	ResponseMs   int64         `json:"responseMs"`
	ErrorMessage string        `json:"errorMessage,omitempty"`
	CheckedAt    time.Time     `json:"checkedAt"`
}

// CheckJob is the scheduler-to-worker message. The worker loads the monitor
// itself so it always checks the current URL and configuration.
type CheckJob struct {
	MonitorID uuid.UUID `json:"monitorId"`
}
