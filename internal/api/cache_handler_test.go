package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFlushCache_All(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &Server{}
	router := gin.New()
	router.POST("/api/cache/flush", server.flushCache)

	payload := FlushCacheRequest{Prefix: "all"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/cache/flush", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	meta, ok := responseBody["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("response does not contain meta object")
	}

	if meta["status"] != "success" {
		t.Fatalf("expected status success, got %v", meta["status"])
	}
}

func TestFlushCache_SpecificPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &Server{}
	router := gin.New()
	router.POST("/api/cache/flush", server.flushCache)

	payload := FlushCacheRequest{Prefix: "tech_stacks:"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/cache/flush", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}
