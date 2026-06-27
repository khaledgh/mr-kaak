package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// TestHealthLive verifies the liveness endpoint responds 200 with the standard
// envelope and does not touch external dependencies (nil DB/Redis are fine).
func TestHealthLive(t *testing.T) {
	e := echo.New()
	h := NewHealth(nil, nil, "test-1.0")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Live(c); err != nil {
		t.Fatalf("Live returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Data struct {
			Status  string `json:"status"`
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Data.Status)
	}
	if body.Data.Version != "test-1.0" {
		t.Errorf("version = %q, want test-1.0", body.Data.Version)
	}
}
