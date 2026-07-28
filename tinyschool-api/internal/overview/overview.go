package overview

import (
	"net/http"

	"tinyschool-api/internal/webjson"
)

type Overview struct {
	Students     int       `json:"students"`
	Classes      int       `json:"classes"`
	Assignments  int       `json:"assignments"`
	Exams        int       `json:"exams"`
	School       Selection `json:"school"`
	AcademicYear Selection `json:"academicYear"`
}

type Selection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/overview", func(w http.ResponseWriter, _ *http.Request) {
		webjson.Write(w, http.StatusOK, map[string]any{"data": Overview{
			Students: 24, Classes: 4, Assignments: 6, Exams: 3,
			School:       Selection{ID: "sch_001", Name: "Tiny School Academy"},
			AcademicYear: Selection{ID: "ay_2026", Name: "2026–27"},
		}})
	})
}
