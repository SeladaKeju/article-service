package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router(nil).ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "{\"status\":\"ok\"}" {
		t.Fatalf("got status %d and body %q", w.Code, w.Body.String())
	}
}
