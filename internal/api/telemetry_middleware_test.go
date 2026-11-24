package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestTelemetryMiddleware_Success(t *testing.T) {
	var mu sync.Mutex
	var receivedEvent TelemetryEvent
	var received bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read body: %v", err)
			return
		}
		if err := json.Unmarshal(body, &receivedEvent); err != nil {
			t.Errorf("failed to unmarshal event: %v", err)
			return
		}
		received = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	router := gin.New()
	router.Use(TelemetryMiddleware(server.URL))
	router.GET("/test/:id", func(c *gin.Context) {
		time.Sleep(5 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test/42", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	// Wait briefly for asynchronous goroutine dispatch
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		mu.Lock()
		got := received
		mu.Unlock()
		if got || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if !received {
		t.Fatal("expected telemetry event to be received by monitor mock")
	}

	if receivedEvent.Source != "go" {
		t.Errorf("expected source 'go', got '%s'", receivedEvent.Source)
	}
	if receivedEvent.Type != "http_request" {
		t.Errorf("expected type 'http_request', got '%s'", receivedEvent.Type)
	}
	if receivedEvent.Endpoint != "/test/:id" {
		t.Errorf("expected endpoint '/test/:id', got '%s'", receivedEvent.Endpoint)
	}
	if receivedEvent.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200, got %d", receivedEvent.StatusCode)
	}
	if receivedEvent.DurationMs <= 0 {
		t.Errorf("expected duration_ms > 0, got %f", receivedEvent.DurationMs)
	}
}

func TestTelemetryMiddleware_Disabled(t *testing.T) {
	router := gin.New()
	router.Use(TelemetryMiddleware(""))
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
}

func TestTelemetryMiddleware_UnreachableServer(t *testing.T) {
	// Point to unreachable port
	router := gin.New()
	router.Use(TelemetryMiddleware("http://127.0.0.1:54321/api/events"))
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200 despite unreachable monitor, got %d", resp.Code)
	}
}

func TestTelemetryMiddleware_SlowServer(t *testing.T) {
	// Server that sleeps longer than telemetryClient timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	router := gin.New()
	router.Use(TelemetryMiddleware(server.URL))
	router.GET("/fast", func(c *gin.Context) {
		c.String(http.StatusOK, "fast")
	})

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	elapsed := time.Since(start)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	// Request must return immediately without waiting for slow monitor
	if elapsed > 50*time.Millisecond {
		t.Errorf("request was blocked by telemetry dispatch, took %v", elapsed)
	}
}

func TestDispatchTelemetry_CustomEvent(t *testing.T) {
	var mu sync.Mutex
	var receivedEvent TelemetryEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		_ = json.NewDecoder(r.Body).Decode(&receivedEvent)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	DispatchTelemetry(server.URL, TelemetryEvent{
		Source: "go",
		Type:   "contact_submit",
		Metadata: map[string]string{
			"category": "inquiry",
		},
	})

	// Wait briefly
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedEvent.Type != "contact_submit" {
		t.Errorf("expected type 'contact_submit', got '%s'", receivedEvent.Type)
	}
	if receivedEvent.Source != "go" {
		t.Errorf("expected source 'go', got '%s'", receivedEvent.Source)
	}
}
