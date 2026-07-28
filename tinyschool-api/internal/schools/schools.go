package schools

import (
	"net/http"
	"strings"

	"tinyschool-api/internal/listquery"
	"tinyschool-api/internal/webjson"
)

type School struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Grades   []string `json:"grades"`
	IsActive bool     `json:"isActive"`
}

var fixtures = []School{
	{ID: "sch_001", Name: "Tiny School Academy", Grades: []string{"Grade 6", "Grade 7", "Grade 8"}, IsActive: true},
	{ID: "sch_002", Name: "Tiny School Primary", Grades: []string{"Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5"}, IsActive: false},
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/schools", list)
}

func list(w http.ResponseWriter, r *http.Request) {
	options, err := listquery.Parse(r.URL.Query(), map[string]bool{"name": true}, "name")
	if err != nil {
		webjson.Error(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	result := listquery.Apply(fixtures, options,
		func(item School, search string) bool {
			return listquery.Contains(search, append([]string{item.Name}, item.Grades...)...)
		},
		func(a, b School, _ string) bool { return strings.ToLower(a.Name) < strings.ToLower(b.Name) },
	)
	webjson.Write(w, http.StatusOK, map[string]any{
		"data": result.Items,
		"meta": map[string]int{"total": result.Total, "page": options.Page, "pageSize": options.PageSize},
	})
}
