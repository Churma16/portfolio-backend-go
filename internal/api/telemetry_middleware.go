package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var telemetryClient = &http.Client{
	Timeout: 100 * time.Millisecond,
}

// TelemetryEvent represents an event payload sent to portfolio-monitor
type TelemetryEvent struct {
	Source     string         `json:"source"`
	Type       string         `json:"type"`
	Endpoint   string         `json:"endpoint,omitempty"`
	StatusCode int            `json:"status_code,omitempty"`
	DurationMs float64        `json:"duration_ms,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// TelemetryMiddleware records request duration and status, sending metrics to portfolio-monitor
func TelemetryMiddleware(monitorURL string) gin.HandlerFunc {
	url := strings.TrimSpace(monitorURL)
	if url == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		durationMs := float64(time.Since(start).Microseconds()) / 1000.0
		status := c.Writer.Status()

		event := TelemetryEvent{
			Source:     "go",
			Type:       "http_request",
			Endpoint:   endpoint,
			StatusCode: status,
			DurationMs: durationMs,
		}

		go DispatchTelemetry(url, event)
	}
}

// DispatchTelemetry sends an event to portfolio-monitor asynchronously without blocking the caller
func DispatchTelemetry(monitorURL string, event TelemetryEvent) {
	url := strings.TrimSpace(monitorURL)
	if url == "" {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := telemetryClient.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
