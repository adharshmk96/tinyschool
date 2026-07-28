package httpdelivery

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/service"
	"tinyschool-api/internal/storage"
	"tinyschool-api/internal/storage/gormsqlite"
	"tinyschool-api/internal/tenancy"
)

func TestSQLiteAuthenticationAndPersistentMutation(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "tinyschool.db")
	store, err := gormsqlite.Open(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AutoMigrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(t.Context()); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := service.New(
		store,
		service.WithJWTSecret([]byte(strings.Repeat("s", 32))),
		service.WithClock(func() time.Time { return time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC) }),
	)
	handler := NewHandler(app, logger)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(
		`{"email":"alex@tinyschool.local","password":"password"}`,
	))
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, login)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResponse.Code, loginResponse.Body.String())
	}
	cookies := loginResponse.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("login cookies = %d", len(cookies))
	}

	overview := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	overview.AddCookie(cookies[0])
	overviewResponse := httptest.NewRecorder()
	handler.ServeHTTP(overviewResponse, overview)
	if overviewResponse.Code != http.StatusOK {
		t.Fatalf("overview status = %d, body = %s", overviewResponse.Code, overviewResponse.Body.String())
	}
	var overviewBody struct {
		Data dto.Overview `json:"data"`
	}
	if err := json.Unmarshal(overviewResponse.Body.Bytes(), &overviewBody); err != nil {
		t.Fatal(err)
	}
	if len(overviewBody.Data.Upcoming) != 5 {
		t.Fatalf("upcoming items = %d, want 5", len(overviewBody.Data.Upcoming))
	}
	if first := overviewBody.Data.Upcoming[0]; first.ID != "asg_006" || first.Date != "2026-07-30" {
		t.Fatalf("first upcoming item = %+v", first)
	}

	create := httptest.NewRequest(http.MethodPost, "/api/v1/schools", strings.NewReader(
		`{"name":"Persistent School","grades":["Grade 1"],"isActive":true}`,
	))
	create.AddCookie(cookies[0])
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, create)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createResponse.Code, createResponse.Body.String())
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := gormsqlite.Open(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	items, total, err := reopened.ListSchools(tenancy.WithUserID(t.Context(), "usr_001"), storage.ListOptions{
		Sort: "name", Order: "asc", Page: 1, PageSize: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(items) != 3 {
		t.Fatalf("persistent schools total = %d, items = %d", total, len(items))
	}
}
