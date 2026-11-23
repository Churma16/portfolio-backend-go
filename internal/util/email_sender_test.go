package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSendNotification(t *testing.T) {
	var received NotificationPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	os.Setenv("NOTIFIER_URL", server.URL)
	defer os.Unsetenv("NOTIFIER_URL")

	err := SendNotification("John Doe", "john@example.com", "Test notification content", "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%s'", received.Name)
	}
	if received.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got '%s'", received.Email)
	}
	if received.Message != "Test notification content" {
		t.Errorf("expected message 'Test notification content', got '%s'", received.Message)
	}
	if received.Source != "go" {
		t.Errorf("expected source 'go', got '%s'", received.Source)
	}
}

func TestSendNotification_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	os.Setenv("NOTIFIER_URL", server.URL)
	defer os.Unsetenv("NOTIFIER_URL")

	err := SendNotification("John Doe", "john@example.com", "Test", "go")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
