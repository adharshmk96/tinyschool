package exams

import (
	"net/http"
	"strings"

	"tinyschool-api/internal/listquery"
	"tinyschool-api/internal/webjson"
)

type Reference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StudentScore struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Grade    string   `json:"grade"`
	Score    *float64 `json:"score"`
	MarkedAt string   `json:"markedAt,omitempty"`
}

type Performance struct {
	AverageScore   float64   `json:"averageScore"`
	CompletionRate int       `json:"completionRate"`
	Completed      int       `json:"completed"`
	Total          int       `json:"total"`
	Standing       string    `json:"standing"`
	Trend          []float64 `json:"trend"`
}

type Exam struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	ExamDate     string         `json:"examDate"`
	TotalScore   float64        `json:"totalScore"`
	Class        Reference      `json:"class"`
	StudentCount int            `json:"studentCount"`
	MarkedCount  int            `json:"markedCount"`
	AverageScore float64        `json:"averageScore"`
	Performance  *Performance   `json:"performance,omitempty"`
	Students     []StudentScore `json:"students,omitempty"`
}

func score(value float64) *float64 { return &value }

var fixtures = []Exam{
	{
		ID: "exam_001", Name: "Mathematics Midterm", Type: "midterm", ExamDate: "2026-08-15", TotalScore: 100,
		Class: Reference{ID: "cls_math7", Name: "Mathematics 7A"}, StudentCount: 3, MarkedCount: 2, AverageScore: 80,
		Performance: &Performance{AverageScore: 80, CompletionRate: 67, Completed: 2, Total: 3, Standing: "On track", Trend: []float64{72, 76, 80}},
		Students: []StudentScore{
			{ID: "stu_001", Name: "Maya Patel", Grade: "Grade 7", Score: score(86), MarkedAt: "2026-08-15T11:00:00Z"},
			{ID: "stu_002", Name: "Noah Williams", Grade: "Grade 7", Score: score(74), MarkedAt: "2026-08-15T11:05:00Z"},
			{ID: "stu_003", Name: "Aarav Shah", Grade: "Grade 7"},
		},
	},
	{
		ID: "exam_002", Name: "Science Quiz", Type: "quiz", ExamDate: "2026-08-20", TotalScore: 30,
		Class: Reference{ID: "cls_sci7", Name: "Science 7A"}, StudentCount: 3, MarkedCount: 0,
		Performance: &Performance{AverageScore: 0, CompletionRate: 0, Completed: 0, Total: 3, Standing: "Not marked", Trend: []float64{}},
		Students:    []StudentScore{{ID: "stu_001", Name: "Maya Patel", Grade: "Grade 7"}, {ID: "stu_002", Name: "Noah Williams", Grade: "Grade 7"}, {ID: "stu_003", Name: "Aarav Shah", Grade: "Grade 7"}},
	},
	{
		ID: "exam_003", Name: "English Assessment", Type: "assessment", ExamDate: "2026-08-24", TotalScore: 50,
		Class: Reference{ID: "cls_eng6", Name: "English 6B"}, StudentCount: 2, MarkedCount: 1, AverageScore: 44,
		Performance: &Performance{AverageScore: 88, CompletionRate: 50, Completed: 1, Total: 2, Standing: "1 of 2 marked", Trend: []float64{72, 78, 84, 88}},
		Students:    []StudentScore{{ID: "stu_004", Name: "Emma Chen", Grade: "Grade 6", Score: score(44), MarkedAt: "2026-08-24T10:30:00Z"}, {ID: "stu_005", Name: "Liam Brown", Grade: "Grade 6"}},
	},
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/exams", list)
	mux.HandleFunc("GET /api/v1/exams/{id}", detail)
}

func list(w http.ResponseWriter, r *http.Request) {
	options, err := listquery.Parse(r.URL.Query(), map[string]bool{
		"name": true, "type": true, "examDate": true, "markedCount": true, "averageScore": true,
	}, "examDate")
	if err != nil {
		webjson.Error(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	result := listquery.Apply(fixtures, options,
		func(item Exam, search string) bool {
			return listquery.Contains(search, item.Name, item.Type, item.Class.Name)
		},
		func(a, b Exam, field string) bool {
			switch field {
			case "name":
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			case "type":
				return a.Type < b.Type
			case "markedCount":
				return a.MarkedCount < b.MarkedCount
			case "averageScore":
				return a.AverageScore < b.AverageScore
			default:
				return a.ExamDate < b.ExamDate
			}
		},
	)
	summaries := make([]Exam, len(result.Items))
	copy(summaries, result.Items)
	for index := range summaries {
		summaries[index].Students = nil
		summaries[index].Performance = nil
	}
	webjson.Write(w, http.StatusOK, map[string]any{
		"data": summaries,
		"meta": map[string]int{"total": result.Total, "page": options.Page, "pageSize": options.PageSize},
	})
}

func detail(w http.ResponseWriter, r *http.Request) {
	for _, item := range fixtures {
		if item.ID == r.PathValue("id") {
			webjson.Write(w, http.StatusOK, map[string]any{"data": item})
			return
		}
	}
	webjson.Error(w, http.StatusNotFound, "not_found", "exam not found")
}
