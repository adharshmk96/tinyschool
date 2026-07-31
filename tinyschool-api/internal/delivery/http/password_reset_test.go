package httpdelivery

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tinyschool-api/internal/service"
	"tinyschool-api/internal/storage/gormsqlite"
)

// resetHandler wires the real SQLite store so the token round-trips through
// storage, and captures the log the reset link is written to.
func resetHandler(t *testing.T) (http.Handler, *bytes.Buffer) {
	t.Helper()
	store, err := gormsqlite.Open(filepath.Join(t.TempDir(), "tinyschool.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(t.Context()); err != nil {
		t.Fatal(err)
	}
	logs := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logs, nil))
	app := service.New(
		store,
		service.WithJWTSecret([]byte(strings.Repeat("s", 32))),
		service.WithClock(func() time.Time { return time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC) }),
		service.WithAppBaseURL("https://school.example/"),
		service.WithLogger(logger),
	)
	return NewHandler(app, logger), logs
}

func postJSON(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	return response
}

// loggedResetToken pulls the token out of the reset link the service logged.
func loggedResetToken(t *testing.T, logs *bytes.Buffer) string {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var entry struct {
			ResetLink string `json:"resetLink"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil || entry.ResetLink == "" {
			continue
		}
		link, err := url.Parse(entry.ResetLink)
		if err != nil {
			t.Fatalf("reset link %q: %v", entry.ResetLink, err)
		}
		if link.Path != "/reset-password" {
			t.Fatalf("reset link path = %q", link.Path)
		}
		return link.Query().Get("token")
	}
	t.Fatalf("no reset link logged in %s", logs.String())
	return ""
}

func TestForgotPasswordLogsLinkAndResetReplacesPassword(t *testing.T) {
	handler, logs := resetHandler(t)

	forgot := postJSON(t, handler, "/api/v1/auth/forgot-password", `{"email":"alex@tinyschool.local"}`)
	if forgot.Code != http.StatusAccepted {
		t.Fatalf("forgot status = %d, body = %s", forgot.Code, forgot.Body.String())
	}
	token := loggedResetToken(t, logs)
	if token == "" {
		t.Fatal("logged reset link carries no token")
	}
	if strings.Contains(logs.String(), "password") && strings.Contains(logs.String(), `"newPassword"`) {
		t.Fatal("log leaked a password")
	}

	body, err := json.Marshal(map[string]string{"token": token, "newPassword": "brand-new-password"})
	if err != nil {
		t.Fatal(err)
	}
	reset := postJSON(t, handler, "/api/v1/auth/reset-password", string(body))
	if reset.Code != http.StatusNoContent {
		t.Fatalf("reset status = %d, body = %s", reset.Code, reset.Body.String())
	}

	old := postJSON(t, handler, "/api/v1/auth/login", `{"email":"alex@tinyschool.local","password":"password"}`)
	if old.Code != http.StatusUnauthorized {
		t.Fatalf("login with old password status = %d", old.Code)
	}
	fresh := postJSON(t, handler, "/api/v1/auth/login", `{"email":"alex@tinyschool.local","password":"brand-new-password"}`)
	if fresh.Code != http.StatusOK {
		t.Fatalf("login with new password status = %d, body = %s", fresh.Code, fresh.Body.String())
	}

	// The token is single use.
	replay := postJSON(t, handler, "/api/v1/auth/reset-password", string(body))
	if replay.Code != http.StatusUnauthorized {
		t.Fatalf("replayed reset status = %d, body = %s", replay.Code, replay.Body.String())
	}
}

func TestForgotPasswordHidesUnknownAccounts(t *testing.T) {
	handler, logs := resetHandler(t)
	response := postJSON(t, handler, "/api/v1/auth/forgot-password", `{"email":"nobody@tinyschool.local"}`)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if strings.Contains(logs.String(), "resetLink") {
		t.Fatalf("issued a reset link for an unknown account: %s", logs.String())
	}
}

func TestResetPasswordRejectsUnknownToken(t *testing.T) {
	handler, _ := resetHandler(t)
	response := postJSON(t, handler, "/api/v1/auth/reset-password",
		`{"token":"not-a-real-token","newPassword":"brand-new-password"}`)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestResetPasswordRejectsShortPassword(t *testing.T) {
	handler, logs := resetHandler(t)
	if code := postJSON(t, handler, "/api/v1/auth/forgot-password", `{"email":"alex@tinyschool.local"}`).Code; code != http.StatusAccepted {
		t.Fatalf("forgot status = %d", code)
	}
	body, err := json.Marshal(map[string]string{"token": loggedResetToken(t, logs), "newPassword": "short"})
	if err != nil {
		t.Fatal(err)
	}
	response := postJSON(t, handler, "/api/v1/auth/reset-password", string(body))
	if response.Code != http.StatusBadRequest && response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
