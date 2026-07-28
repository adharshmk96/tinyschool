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

	"tinyschool-api/internal/service"
	"tinyschool-api/internal/storage/gormsqlite"
)

func adminTestHandler(t *testing.T) http.Handler {
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
	return NewHandler(service.New(store, service.WithJWTSecret([]byte(strings.Repeat("s", 32)))), logger)
}

func do(t *testing.T, handler http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func adminCookie(t *testing.T, response *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == adminSessionCookieName {
			return cookie
		}
	}
	t.Fatalf("no admin cookie in response: %s", response.Body.String())
	return nil
}

func TestAdminSetupThenLoginAndBlockUser(t *testing.T) {
	handler := adminTestHandler(t)

	status := do(t, handler, http.MethodGet, "/api/v1/admin/status", "", nil)
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), `"adminExists":false`) {
		t.Fatalf("status = %d, body = %s", status.Code, status.Body.String())
	}

	setup := do(t, handler, http.MethodPost, "/api/v1/admin/setup",
		`{"name":"Root","email":"root@tinyschool.local","password":"password123"}`, nil)
	if setup.Code != http.StatusCreated {
		t.Fatalf("setup status = %d, body = %s", setup.Code, setup.Body.String())
	}
	cookie := adminCookie(t, setup)

	again := do(t, handler, http.MethodPost, "/api/v1/admin/setup",
		`{"name":"Root","email":"other@tinyschool.local","password":"password123"}`, nil)
	if again.Code != http.StatusConflict {
		t.Fatalf("second setup status = %d, body = %s", again.Code, again.Body.String())
	}

	users := do(t, handler, http.MethodGet, "/api/v1/admin/users?sort=email&order=asc", "", cookie)
	if users.Code != http.StatusOK {
		t.Fatalf("users status = %d, body = %s", users.Code, users.Body.String())
	}
	var listed struct {
		Data []struct {
			ID      string `json:"id"`
			Email   string `json:"email"`
			Role    string `json:"role"`
			Blocked bool   `json:"blocked"`
		} `json:"data"`
	}
	if err := json.Unmarshal(users.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	var seededID string
	for _, user := range listed.Data {
		if user.Role != "admin" {
			seededID = user.ID
		}
	}
	if seededID == "" {
		t.Fatalf("no seeded user in %s", users.Body.String())
	}

	block := do(t, handler, http.MethodPost, "/api/v1/admin/users/"+seededID+"/block", "", cookie)
	if block.Code != http.StatusOK || !strings.Contains(block.Body.String(), `"blocked":true`) {
		t.Fatalf("block status = %d, body = %s", block.Code, block.Body.String())
	}

	login := do(t, handler, http.MethodPost, "/api/v1/auth/login",
		`{"email":"alex@tinyschool.local","password":"password"}`, nil)
	if login.Code != http.StatusUnauthorized {
		t.Fatalf("blocked login status = %d, body = %s", login.Code, login.Body.String())
	}

	unblock := do(t, handler, http.MethodPost, "/api/v1/admin/users/"+seededID+"/unblock", "", cookie)
	if unblock.Code != http.StatusOK || !strings.Contains(unblock.Body.String(), `"blocked":false`) {
		t.Fatalf("unblock status = %d, body = %s", unblock.Code, unblock.Body.String())
	}
	if login := do(t, handler, http.MethodPost, "/api/v1/auth/login",
		`{"email":"alex@tinyschool.local","password":"password"}`, nil); login.Code != http.StatusOK {
		t.Fatalf("unblocked login status = %d, body = %s", login.Code, login.Body.String())
	}

	adminLogin := do(t, handler, http.MethodPost, "/api/v1/admin/login",
		`{"email":"alex@tinyschool.local","password":"password"}`, nil)
	if adminLogin.Code != http.StatusUnauthorized {
		t.Fatalf("non-admin console login status = %d", adminLogin.Code)
	}
}

func TestAdminDeleteUserRemovesOwnedData(t *testing.T) {
	handler := adminTestHandler(t)
	setup := do(t, handler, http.MethodPost, "/api/v1/admin/setup",
		`{"name":"Root","email":"root@tinyschool.local","password":"password123"}`, nil)
	cookie := adminCookie(t, setup)

	login := do(t, handler, http.MethodPost, "/api/v1/auth/login",
		`{"email":"alex@tinyschool.local","password":"password"}`, nil)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body.String())
	}
	session := login.Result().Cookies()[0]
	schools := do(t, handler, http.MethodGet, "/api/v1/schools", "", session)
	if schools.Code != http.StatusOK || strings.Contains(schools.Body.String(), `"data":[]`) {
		t.Fatalf("seeded schools status = %d, body = %s", schools.Code, schools.Body.String())
	}

	var listed struct {
		Data []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"data"`
	}
	users := do(t, handler, http.MethodGet, "/api/v1/admin/users", "", cookie)
	if err := json.Unmarshal(users.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	var seededID string
	for _, user := range listed.Data {
		if user.Email == "alex@tinyschool.local" {
			seededID = user.ID
		}
	}
	if seededID == "" {
		t.Fatalf("seeded user missing from %s", users.Body.String())
	}

	if response := do(t, handler, http.MethodDelete, "/api/v1/admin/users/"+seededID, "", cookie); response.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, body = %s", response.Code, response.Body.String())
	}
	// The account, its live session and everything it owned are gone.
	if response := do(t, handler, http.MethodGet, "/api/v1/schools", "", session); response.Code != http.StatusUnauthorized {
		t.Fatalf("session after delete status = %d, body = %s", response.Code, response.Body.String())
	}
	if response := do(t, handler, http.MethodPost, "/api/v1/auth/login",
		`{"email":"alex@tinyschool.local","password":"password"}`, nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("login after delete status = %d, body = %s", response.Code, response.Body.String())
	}
	remaining := do(t, handler, http.MethodGet, "/api/v1/admin/users", "", cookie)
	if strings.Contains(remaining.Body.String(), "alex@tinyschool.local") {
		t.Fatalf("deleted user still listed: %s", remaining.Body.String())
	}
	// Re-registering with the same email proves the owned rows were released.
	if response := do(t, handler, http.MethodPost, "/api/v1/auth/register",
		`{"name":"Alex","email":"alex@tinyschool.local","password":"password"}`, nil); response.Code != http.StatusCreated {
		t.Fatalf("re-register status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAdminCannotDeleteSelfOrLastAdmin(t *testing.T) {
	handler := adminTestHandler(t)
	setup := do(t, handler, http.MethodPost, "/api/v1/admin/setup",
		`{"name":"Root","email":"root@tinyschool.local","password":"password123"}`, nil)
	cookie := adminCookie(t, setup)

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(setup.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if response := do(t, handler, http.MethodDelete, "/api/v1/admin/users/"+created.Data.ID, "", cookie); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("self delete status = %d, body = %s", response.Code, response.Body.String())
	}

	second := do(t, handler, http.MethodPost, "/api/v1/admin/admins",
		`{"name":"Second","email":"second@tinyschool.local","password":"password123"}`, cookie)
	if second.Code != http.StatusCreated {
		t.Fatalf("add admin status = %d, body = %s", second.Code, second.Body.String())
	}
	if err := json.Unmarshal(second.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	// A second administrator exists, so deleting one is allowed.
	if response := do(t, handler, http.MethodDelete, "/api/v1/admin/users/"+created.Data.ID, "", cookie); response.Code != http.StatusNoContent {
		t.Fatalf("delete other admin status = %d, body = %s", response.Code, response.Body.String())
	}
	if response := do(t, handler, http.MethodDelete, "/api/v1/admin/users/usr_missing", "", cookie); response.Code != http.StatusNotFound {
		t.Fatalf("missing user delete status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestOnlyAnAdminCanAddAnotherAdmin(t *testing.T) {
	handler := adminTestHandler(t)
	body := `{"name":"Second","email":"second@tinyschool.local","password":"password123"}`

	if response := do(t, handler, http.MethodPost, "/api/v1/admin/admins", body, nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous add status = %d, body = %s", response.Code, response.Body.String())
	}

	setup := do(t, handler, http.MethodPost, "/api/v1/admin/setup",
		`{"name":"Root","email":"root@tinyschool.local","password":"password123"}`, nil)
	if setup.Code != http.StatusCreated {
		t.Fatalf("setup status = %d, body = %s", setup.Code, setup.Body.String())
	}
	cookie := adminCookie(t, setup)

	// The public bootstrap route stays closed once an administrator exists.
	if response := do(t, handler, http.MethodPost, "/api/v1/admin/setup", body, nil); response.Code != http.StatusConflict {
		t.Fatalf("setup after bootstrap status = %d, body = %s", response.Code, response.Body.String())
	}
	// A school session must not be able to mint an administrator either.
	login := do(t, handler, http.MethodPost, "/api/v1/auth/login",
		`{"email":"alex@tinyschool.local","password":"password"}`, nil)
	forged := &http.Cookie{Name: adminSessionCookieName, Value: login.Result().Cookies()[0].Value}
	if response := do(t, handler, http.MethodPost, "/api/v1/admin/admins", body, forged); response.Code != http.StatusUnauthorized {
		t.Fatalf("school session add status = %d, body = %s", response.Code, response.Body.String())
	}

	added := do(t, handler, http.MethodPost, "/api/v1/admin/admins", body, cookie)
	if added.Code != http.StatusCreated || !strings.Contains(added.Body.String(), `"role":"admin"`) {
		t.Fatalf("add status = %d, body = %s", added.Code, added.Body.String())
	}
	if duplicate := do(t, handler, http.MethodPost, "/api/v1/admin/admins", body, cookie); duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate email status = %d, body = %s", duplicate.Code, duplicate.Body.String())
	}
	if response := do(t, handler, http.MethodPost, "/api/v1/admin/login",
		`{"email":"second@tinyschool.local","password":"password123"}`, nil); response.Code != http.StatusOK {
		t.Fatalf("new admin login status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAdminRoutesRequireAdminSession(t *testing.T) {
	handler := adminTestHandler(t)
	if response := do(t, handler, http.MethodGet, "/api/v1/admin/users", "", nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	login := do(t, handler, http.MethodPost, "/api/v1/auth/login",
		`{"email":"alex@tinyschool.local","password":"password"}`, nil)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body.String())
	}
	// A school session cookie renamed onto the admin cookie must not pass.
	forged := &http.Cookie{Name: adminSessionCookieName, Value: login.Result().Cookies()[0].Value}
	if response := do(t, handler, http.MethodGet, "/api/v1/admin/users", "", forged); response.Code != http.StatusUnauthorized {
		t.Fatalf("forged status = %d, body = %s", response.Code, response.Body.String())
	}
}
