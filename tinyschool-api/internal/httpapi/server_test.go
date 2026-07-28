package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler() http.Handler {
	return NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestRequiredEndpoints(t *testing.T) {
	endpoints := []string{
		"/health",
		"/api/v1/me",
		"/api/v1/overview",
		"/api/v1/schools",
		"/api/v1/academic-years",
		"/api/v1/academic-years/ay_2026",
		"/api/v1/classes",
		"/api/v1/classes/cls_math7",
		"/api/v1/students",
		"/api/v1/students/stu_001",
		"/api/v1/assignments",
		"/api/v1/assignments/asg_001",
		"/api/v1/exams",
		"/api/v1/exams/exam_001",
	}
	handler := newTestHandler()
	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, endpoint, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body)
			}
			if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
				t.Fatalf("Content-Type = %q", contentType)
			}
			var body any
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
		})
	}
}

func TestRequestedDetailFixturesIncludeTabData(t *testing.T) {
	tests := []struct {
		path        string
		collections []string
	}{
		{path: "/api/v1/students/stu_006", collections: []string{"assignments", "exams", "behaviour", "notes"}},
		{path: "/api/v1/classes/cls_sci7", collections: []string{"assignments", "exams", "students"}},
		{path: "/api/v1/assignments/asg_005", collections: []string{"assignees"}},
		{path: "/api/v1/exams/exam_003", collections: []string{"students"}},
	}

	handler := newTestHandler()
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body)
			}

			var body struct {
				Data map[string]any `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if _, ok := body.Data["performance"]; test.path != "/api/v1/assignments/asg_005" && !ok {
				t.Fatal("performance data is missing")
			}
			for _, field := range test.collections {
				items, ok := body.Data[field].([]any)
				if !ok || len(items) == 0 {
					t.Fatalf("%s data is missing or empty", field)
				}
			}
		})
	}
}

func TestCollectionSearchSortAndPagination(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/students?search=grade+7&sort=name&order=desc&page=2&pageSize=1", nil)
	response := httptest.NewRecorder()
	newTestHandler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body)
	}
	var body struct {
		Data []struct {
			FullName string `json:"fullName"`
		} `json:"data"`
		Meta struct {
			Total int `json:"total"`
			Page  int `json:"page"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Meta.Total != 3 || body.Meta.Page != 2 || len(body.Data) != 1 || body.Data[0].FullName != "Maya Patel" {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestInvalidQueryReturnsJSONError(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/classes?page=zero", nil)
	response := httptest.NewRecorder()
	newTestHandler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "invalid_query" {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}

func TestMissingDetailAndUnknownRouteReturnJSON404(t *testing.T) {
	for _, endpoint := range []string{"/api/v1/students/missing", "/not-real"} {
		request := httptest.NewRequest(http.MethodGet, endpoint, nil)
		response := httptest.NewRecorder()
		newTestHandler().ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, body = %s", endpoint, response.Code, response.Body)
		}
		if !strings.Contains(response.Body.String(), `"code":"not_found"`) {
			t.Fatalf("%s body = %s", endpoint, response.Body)
		}
	}
}

func TestCORSPreflight(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/students", nil)
	response := httptest.NewRecorder()
	newTestHandler().ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
}
