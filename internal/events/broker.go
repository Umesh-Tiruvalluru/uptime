// Package events fans NATS check-completion messages out to SSE clients.
package events

import (
	"encoding/json"
	"sync"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/natssubj"
	"github.com/nats-io/nats.go"
)

type Broker struct {
	mu      sync.RWMutex
	clients map[chan models.CheckResult]struct{}
}

func NewBroker(nc *nats.Conn) (*Broker, error) {
	b := &Broker{clients: make(map[chan models.CheckResult]struct{})}
	_, err := nc.Subscribe(natssubj.ChecksCompleted, func(msg *nats.Msg) {
		var result models.CheckResult
		if err := json.Unmarshal(msg.Data, &result); err == nil {
			b.publish(result)
		}
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Broker) Subscribe() (<-chan models.CheckResult, func()) {
	client := make(chan models.CheckResult, 16)
	b.mu.Lock()
	b.clients[client] = struct{}{}
	b.mu.Unlock()

	return client, func() {
		b.mu.Lock()
		delete(b.clients, client)
		b.mu.Unlock()
	}
}

func (b *Broker) publish(result models.CheckResult) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for client := range b.clients {
		select {
		case client <- result:
		default: // Slow clients must not block the real-time event pipeline.
		}
	}
}
