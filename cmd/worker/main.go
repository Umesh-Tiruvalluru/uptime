package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/emailer"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/natssubj"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

const checkTimeout = 10 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL, nats.Name("monitoring-worker"))
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer nc.Close()

	repo := repository.NewRepository(pool)
	client := &http.Client{Timeout: checkTimeout}
	alerter := emailer.NewResendEmailer()
	if alerter.Enabled() {
		log.Printf("worker: emailer configured (from=%s)", alerter.FromEmail)
	} else {
		log.Printf("worker: RESEND_API_KEY not set — downtime alerts will be logged to stdout")
	}

	_, err = nc.QueueSubscribe(natssubj.ChecksPerform, "monitoring-workers", func(msg *nats.Msg) {
		handleJob(repo, nc, client, alerter, msg)
	})
	if err != nil {
		log.Fatalf("subscribe to check jobs: %v", err)
	}
	if err := nc.Flush(); err != nil {
		log.Fatalf("initialize NATS subscription: %v", err)
	}

	log.Printf("worker listening on %s", natssubj.ChecksPerform)
	<-ctx.Done()
}

func handleJob(repo *repository.Repository, nc *nats.Conn, client *http.Client, alerter emailer.Emailer, msg *nats.Msg) {
	var job models.CheckJob
	if err := json.Unmarshal(msg.Data, &job); err != nil || job.MonitorID.String() == "00000000-0000-0000-0000-000000000000" {
		log.Printf("worker: invalid check job: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout+5*time.Second)
	defer cancel()
	monitor, err := repo.GetMonitorByID(ctx, job.MonitorID)
	if err != nil {
		log.Printf("worker: load monitor %s: %v", job.MonitorID, err)
		return
	}

	result := performHTTPCheck(ctx, client, monitor.ID, monitor.URL)
	if err := repo.RecordCheck(ctx, &result); err != nil {
		log.Printf("worker: save check for %s: %v", monitor.ID, err)
		return
	}

	if result.Status == models.StatusDown {
		notifyOwnerOfDowntime(ctx, repo, alerter, monitor, result)
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("worker: encode completion for %s: %v", monitor.ID, err)
		return
	}
	if err := nc.Publish(natssubj.ChecksCompleted, data); err != nil {
		log.Printf("worker: publish completion for %s: %v", monitor.ID, err)
	}
}

// notifyOwnerOfDowntime resolves the monitor owner's email and sends the alert.
func notifyOwnerOfDowntime(ctx context.Context, repo *repository.Repository, alerter emailer.Emailer, monitor *models.Monitor, result models.CheckResult) {
	owner, err := repo.GetUserByID(ctx, monitor.UserID)
	if err != nil {
		log.Printf("worker: load owner %s for monitor %s: %v", monitor.UserID, monitor.ID, err)
		return
	}
	if owner.Email == "" {
		log.Printf("worker: owner %s has no email — skipping alert for monitor %s", owner.ID, monitor.ID)
		return
	}

	detail := emailer.AlertDetail{
		MonitorName: monitor.Name,
		MonitorURL:  monitor.URL,
		StatusCode:  result.StatusCode,
		Error:       result.ErrorMessage,
		CheckedAt:   result.CheckedAt,
	}
	if err := alerter.SendDowntimeAlert(ctx, owner.Email, detail); err != nil {
		log.Printf("worker: send downtime alert for monitor %s: %v", monitor.ID, err)
	}
}

func performHTTPCheck(ctx context.Context, client *http.Client, monitorID uuid.UUID, targetURL string) models.CheckResult {
	started := time.Now()
	result := models.CheckResult{
		MonitorID: monitorID,
		Status:    models.StatusDown,
		CheckedAt: started.UTC(),
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}
	response, err := client.Do(request)
	result.ResponseMs = time.Since(started).Milliseconds()
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}
	defer response.Body.Close()

	result.StatusCode = response.StatusCode
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusBadRequest {
		result.Status = models.StatusUp
		return result
	}
	result.ErrorMessage = fmt.Sprintf("unexpected HTTP status: %d", response.StatusCode)
	return result
}
