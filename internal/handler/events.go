package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/events"
	"github.com/gin-gonic/gin"
)

func SSE(broker *events.Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")
		c.Status(http.StatusOK)
		flusher.Flush()

		results, unsubscribe := broker.Subscribe()
		defer unsubscribe()
		for {
			select {
			case <-c.Request.Context().Done():
				return
			case result := <-results:
				data, err := json.Marshal(result)
				if err != nil {
					continue
				}
				fmt.Fprintf(c.Writer, "event: check.completed\ndata: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}
