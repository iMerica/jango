package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iMerica/jango/rest"
)

func TestJSONRendererAndAPIResponse(t *testing.T) {
	resp := rest.NewAPIResponse(map[string]interface{}{"ok": true}, http.StatusCreated)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if resp.StatusCode() != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode())
	}
	if err := resp.WriteTo(rec, req); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected response code 201, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json, got %q", rec.Header().Get("Content-Type"))
	}
	var body map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body was not JSON: %v", err)
	}
	if !body["ok"] {
		t.Fatal("expected ok=true")
	}
}

func TestAPIResponseRendererErrorsSurface(t *testing.T) {
	resp := rest.NewAPIResponse(map[string]interface{}{"bad": make(chan int)}, http.StatusOK)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := resp.WriteTo(rec, req); err == nil {
		t.Fatal("expected renderer error")
	}
}
