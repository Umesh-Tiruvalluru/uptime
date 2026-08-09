// Package scheduler periodically turns due monitors into worker jobs.
package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/natssubj"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/nats-io/nats.go"
)

const (
	defaultInterval = 15 * time.Second
	batchSize       = 100
)

// Start runs the scheduler until ctx is cancelled. It executes once at startup
// so newly-created monitors do not wait for the first ticker interval.
func Start(ctx context.Context, repo *repository.Repository, nc *nats.Conn) {
	dispatch(ctx, repo, nc)

	ticker := time.NewTicker(defaultInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dispatch(ctx, repo, nc)
		}
	}
}

func dispatch(ctx context.Context, repo *repository.Repository, nc *nats.Conn) {
	monitors, err := repo.GetDueMonitors(ctx, time.Now(), batchSize)
	if err != nil {
		log.Printf("scheduler: find due monitors: %v", err)
		return
	}

	for _, monitor := range monitors {
		data, err := json.Marshal(models.CheckJob{MonitorID: monitor.ID})
		if err != nil {
			log.Printf("scheduler: encode job for %s: %v", monitor.ID, err)
			continue
		}
		if err := nc.Publish(natssubj.ChecksPerform, data); err != nil {
			log.Printf("scheduler: publish job for %s: %v", monitor.ID, err)
		}
	}
	if len(monitors) > 0 {
		if err := nc.FlushTimeout(2 * time.Second); err != nil {
			log.Printf("scheduler: flush jobs: %v", err)
		}
	}
}
