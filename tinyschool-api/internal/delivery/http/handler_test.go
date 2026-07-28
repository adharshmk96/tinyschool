package httpdelivery

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tinyschool-api/internal/service"
	"tinyschool-api/internal/storage"
)

type testStore struct {
	storage.Storage
	pingError error
}

func (s testStore) Ping(context.Context) error {
	return s.pingError
}

func testHandler(store storage.Storage) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(service.New(store, service.WithJWTSecret([]byte(strings.Repeat("s", 32)))), logger)
}

func TestHealthDoesNotRequireStorage(t *testing.T) {
	response := httptest.NewRecorder()
	testHandler(testStore{pingError: errors.New("offline")}).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/health", nil),
	)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestReadyUsesStoragePing(t *testing.T) {
	response := httptest.NewRecorder()
	testHandler(testStore{pingError: errors.New("offline")}).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/ready", nil),
	)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestProtectedRouteRequiresAuthentication(t *testing.T) {
	response := httptest.NewRecorder()
	testHandler(testStore{}).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/students", nil),
	)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
