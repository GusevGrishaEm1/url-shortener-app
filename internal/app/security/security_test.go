package security

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockShortenerService struct{}

func (m *mockShortenerService) GetUserID(context.Context) int {
	return 666
}

func TestRequiredUserID(t *testing.T) {
	service := &mockShortenerService{}
	securityMiddleware := NewSecurityMiddleware(service)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	securityMiddleware.RequiredUserID(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestSecurity(t *testing.T) {
	service := &mockShortenerService{}
	securityMiddleware := NewSecurityMiddleware(service)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	securityMiddleware.Security(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
