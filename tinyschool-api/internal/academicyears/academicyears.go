package academicyears

import (
	"net/http"
	"strings"

	"tinyschool-api/internal/listquery"
	"tinyschool-api/internal/webjson"
)

type Segment struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	DurationDays int    `json:"durationDays"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
}

type AcademicYear struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	StartDate    string    `json:"startDate"`
	EndDate      string    `json:"endDate"`
	DurationDays int       `json:"durationDays"`
	IsCurrent    bool      `json:"isCurrent"`
	Segments     []Segment `json:"segments"`
}

var fixtures = []AcademicYear{
	{
		ID: "ay_2026", Name: "2026–27", StartDate: "2026-06-01", EndDate: "2027-03-31", DurationDays: 304, IsCurrent: true,
		Segments: []Segment{
			{ID: "seg_001", Name: "Term 1", Type: "term", DurationDays: 120, StartDate: "2026-06-01", EndDate: "2026-09-28"},
			{ID: "seg_002", Name: "Autumn Break", Type: "vacation", DurationDays: 10, StartDate: "2026-09-29", EndDate: "2026-10-08"},
			{ID: "seg_003", Name: "Term 2", Type: "term", DurationDays: 174, StartDate: "2026-10-09", EndDate: "2027-03-31"},
		},
	},
	{
		ID: "ay_2025", Name: "2025–26", StartDate: "2025-06-02", EndDate: "2026-03-31", DurationDays: 303, IsCurrent: false,
		Segments: []Segment{
			{ID: "seg_004", Name: "Term 1", Type: "term", DurationDays: 151, StartDate: "2025-06-02", EndDate: "2025-10-30"},
			{ID: "seg_005", Name: "Term 2", Type: "term", DurationDays: 152, StartDate: "2025-10-31", EndDate: "2026-03-31"},
		},
	},
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/academic-years", list)
	mux.HandleFunc("GET /api/v1/academic-years/{id}", detail)
}

func list(w http.ResponseWriter, r *http.Request) {
	options, err := listquery.Parse(r.URL.Query(), map[string]bool{
		"name": true, "startDate": true, "durationDays": true,
	}, "startDate")
	if err != nil {
		webjson.Error(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	result := listquery.Apply(fixtures, options,
		func(item AcademicYear, search string) bool {
			return listquery.Contains(search, item.Name, item.StartDate)
		},
		func(a, b AcademicYear, field string) bool {
			switch field {
			case "durationDays":
				return a.DurationDays < b.DurationDays
			case "name":
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			default:
				return a.StartDate < b.StartDate
			}
		},
	)
	webjson.Write(w, http.StatusOK, map[string]any{
		"data": result.Items,
		"meta": map[string]int{"total": result.Total, "page": options.Page, "pageSize": options.PageSize},
	})
}

func detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, item := range fixtures {
		if item.ID == id {
			webjson.Write(w, http.StatusOK, map[string]any{"data": item})
			return
		}
	}
	webjson.Error(w, http.StatusNotFound, "not_found", "academic year not found")
}
