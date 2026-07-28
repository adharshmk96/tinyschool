package httpdelivery

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONIsStrict(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty", body: ""},
		{name: "unknown field", body: `{"name":"Ada","extra":true}`},
		{name: "trailing value", body: `{"name":"Ada"} {}`},
		{name: "malformed", body: `{"name":`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			var destination struct {
				Name string `json:"name"`
			}
			if err := decodeJSON(response, request, &destination); err == nil {
				t.Fatal("decodeJSON() error = nil")
			}
		})
	}
}

func TestDecodeJSONAcceptsSingleKnownObject(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ada"}`))
	response := httptest.NewRecorder()
	var destination struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(response, request, &destination); err != nil {
		t.Fatal(err)
	}
	if destination.Name != "Ada" {
		t.Fatalf("Name = %q", destination.Name)
	}
}

func TestCORSSupportsMutationMethodsAndCredentials(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()

	cors(next).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "PATCH") || !strings.Contains(got, "DELETE") {
		t.Fatalf("Access-Control-Allow-Methods = %q", got)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("credentialed origin was not reflected")
	}
	if response.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("credentials are not allowed")
	}
}

func TestRecovererReturnsJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := recoverer(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}
