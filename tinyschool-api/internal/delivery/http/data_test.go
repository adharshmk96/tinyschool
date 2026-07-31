package httpdelivery

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"tinyschool-api/internal/dataio"
	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/service"
	"tinyschool-api/internal/storage/gormsqlite"
)

// TestDataExportImportRoundTrip exports a seeded workspace, imports the export
// back over itself, and checks the workspace still holds the same counts.
func TestDataExportImportRoundTrip(t *testing.T) {
	for _, format := range []struct{ query, filename string }{
		{"xlsx", "tinyschool-export.xlsx"},
		{"csv", "tinyschool-export.zip"},
	} {
		t.Run(format.query, func(t *testing.T) {
			handler, cookie := seededSession(t)

			exported := send(t, handler, cookie, httptest.NewRequest(http.MethodGet, "/api/v1/me/data/export?format="+format.query, nil))
			if exported.Code != http.StatusOK {
				t.Fatalf("export status = %d, body = %s", exported.Code, exported.Body.String())
			}
			file := exported.Body.Bytes()
			if len(file) == 0 {
				t.Fatal("export returned an empty file")
			}

			before := overviewOf(t, handler, cookie)
			summary := importFile(t, handler, cookie, format.filename, file)
			if summary.Schools == 0 || summary.Students == 0 || summary.Classes == 0 {
				t.Fatalf("import summary looks empty: %+v", summary)
			}
			after := overviewOf(t, handler, cookie)
			if before != after {
				t.Fatalf("overview changed across the round trip: before %+v, after %+v", before, after)
			}

			// The re-imported rows must carry fresh ids, not the exported ones.
			second, err := dataio.Decode(format.filename, file)
			if err != nil {
				t.Fatal(err)
			}
			reexported := send(t, handler, cookie, httptest.NewRequest(http.MethodGet, "/api/v1/me/data/export?format="+format.query, nil))
			roundTripped, err := dataio.Decode(format.filename, reexported.Body.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if len(second.Students) != len(roundTripped.Students) {
				t.Fatalf("students after import = %d, want %d", len(roundTripped.Students), len(second.Students))
			}
			if second.Students[0].ID == roundTripped.Students[0].ID {
				t.Fatal("imported rows kept their file ids; ids should be regenerated")
			}
		})
	}
}

func TestImportRejectsBrokenReferences(t *testing.T) {
	handler, cookie := seededSession(t)
	csvFile := "id,schoolId,firstName,lastName\nstu_1,sch_missing,Ada,Lovelace\n"

	response := importResponse(t, handler, cookie, "students.csv", []byte(csvFile))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "not listed in the schools sheet") {
		t.Fatalf("unexpected error body: %s", response.Body.String())
	}
	// A rejected import must leave the workspace untouched.
	if overviewOf(t, handler, cookie).Students == 0 {
		t.Fatal("students were cleared by a rejected import")
	}
}

func TestImportRejectsUnknownFileType(t *testing.T) {
	handler, cookie := seededSession(t)
	response := importResponse(t, handler, cookie, "data.json", []byte("{}"))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func seededSession(t *testing.T) (http.Handler, *http.Cookie) {
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
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(service.New(store, service.WithJWTSecret([]byte(strings.Repeat("s", 32)))), logger)

	login := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(
		`{"email":"alex@tinyschool.local","password":"password"}`,
	))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, login)
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login returned no session cookie")
	}
	return handler, cookies[0]
}

func send(t *testing.T, handler http.Handler, cookie *http.Cookie, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func overviewOf(t *testing.T, handler http.Handler, cookie *http.Cookie) (counts struct{ Students, Classes, Assignments, Exams int }) {
	t.Helper()
	response := send(t, handler, cookie, httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("overview status = %d, body = %s", response.Code, response.Body.String())
	}
	var body struct {
		Data dto.Overview `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	counts.Students = body.Data.Students
	counts.Classes = body.Data.Classes
	counts.Assignments = body.Data.Assignments
	counts.Exams = body.Data.Exams
	return counts
}

func importResponse(t *testing.T, handler http.Handler, cookie *http.Cookie, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/me/data/import", &body)
	request.Header.Set("Content-Type", form.FormDataContentType())
	return send(t, handler, cookie, request)
}

func importFile(t *testing.T, handler http.Handler, cookie *http.Cookie, filename string, content []byte) dto.ImportSummary {
	t.Helper()
	response := importResponse(t, handler, cookie, filename, content)
	if response.Code != http.StatusOK {
		t.Fatalf("import status = %d, body = %s", response.Code, response.Body.String())
	}
	var body struct {
		Data dto.ImportSummary `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Data
}
