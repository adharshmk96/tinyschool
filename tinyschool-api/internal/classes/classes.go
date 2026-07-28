package classes

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

type Performance struct {
	AverageScore   float64   `json:"averageScore"`
	CompletionRate int       `json:"completionRate"`
	Completed      int       `json:"completed"`
	Total          int       `json:"total"`
	Standing       string    `json:"standing"`
	Trend          []float64 `json:"trend"`
}

type Class struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Subject      string       `json:"subject"`
	Grade        string       `json:"grade"`
	Description  string       `json:"description"`
	StudentCount int          `json:"studentCount"`
	AverageScore float64      `json:"averageScore"`
	Performance  *Performance `json:"performance,omitempty"`
	Students     []Reference  `json:"students,omitempty"`
	Assignments  []Reference  `json:"assignments,omitempty"`
	Exams        []Reference  `json:"exams,omitempty"`
}

var fixtures = []Class{
	{
		ID: "cls_math7", Name: "Mathematics 7A", Subject: "Mathematics", Grade: "Grade 7",
		Description: "Core mathematics with an emphasis on algebra and geometry.", StudentCount: 8, AverageScore: 82.4,
		Performance: &Performance{AverageScore: 82.4, CompletionRate: 84, Completed: 21, Total: 25, Standing: "2nd of 4 classes", Trend: []float64{74, 77, 79, 80, 82, 82.4}},
		Students:    []Reference{{ID: "stu_001", Name: "Maya Patel"}, {ID: "stu_002", Name: "Noah Williams"}, {ID: "stu_003", Name: "Aarav Shah"}},
		Assignments: []Reference{{ID: "asg_001", Name: "Algebra Practice"}, {ID: "asg_004", Name: "Geometry Project"}},
		Exams:       []Reference{{ID: "exam_001", Name: "Mathematics Midterm"}},
	},
	{
		ID: "cls_sci7", Name: "Science 7A", Subject: "Science", Grade: "Grade 7",
		Description: "Hands-on life and physical sciences.", StudentCount: 8, AverageScore: 79.8,
		Performance: &Performance{AverageScore: 79.8, CompletionRate: 78, Completed: 18, Total: 23, Standing: "3rd of 4 classes", Trend: []float64{69, 73, 71, 76, 78, 80}},
		Students:    []Reference{{ID: "stu_001", Name: "Maya Patel"}, {ID: "stu_002", Name: "Noah Williams"}, {ID: "stu_003", Name: "Aarav Shah"}},
		Assignments: []Reference{{ID: "asg_002", Name: "Plant Cell Model"}},
		Exams:       []Reference{{ID: "exam_002", Name: "Science Quiz"}},
	},
	{
		ID: "cls_eng6", Name: "English 6B", Subject: "English", Grade: "Grade 6",
		Description: "Reading comprehension and creative writing.", StudentCount: 6, AverageScore: 84.1,
		Students:    []Reference{{ID: "stu_004", Name: "Emma Chen"}, {ID: "stu_005", Name: "Liam Brown"}},
		Assignments: []Reference{{ID: "asg_003", Name: "Short Story Draft"}},
		Exams:       []Reference{{ID: "exam_003", Name: "English Assessment"}},
	},
	{
		ID: "cls_hist8", Name: "History 8A", Subject: "History", Grade: "Grade 8",
		Description: "World history through primary sources.", StudentCount: 2, AverageScore: 77.5,
		Students:    []Reference{{ID: "stu_006", Name: "Olivia Martin"}},
		Assignments: []Reference{{ID: "asg_005", Name: "Ancient Civilizations Essay"}},
	},
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/classes", list)
	mux.HandleFunc("GET /api/v1/classes/{id}", detail)
}

func list(w http.ResponseWriter, r *http.Request) {
	options, err := listquery.Parse(r.URL.Query(), map[string]bool{
		"name": true, "subject": true, "grade": true, "studentCount": true,
	}, "name")
	if err != nil {
		webjson.Error(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	result := listquery.Apply(fixtures, options,
		func(item Class, search string) bool {
			return listquery.Contains(search, item.Name, item.Subject, item.Grade, item.Description)
		},
		func(a, b Class, field string) bool {
			switch field {
			case "subject":
				return strings.ToLower(a.Subject) < strings.ToLower(b.Subject)
			case "grade":
				return strings.ToLower(a.Grade) < strings.ToLower(b.Grade)
			case "studentCount":
				return a.StudentCount < b.StudentCount
			default:
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
		},
	)
	summaries := make([]Class, len(result.Items))
	copy(summaries, result.Items)
	for index := range summaries {
		summaries[index].Students = nil
		summaries[index].Assignments = nil
		summaries[index].Exams = nil
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
	webjson.Error(w, http.StatusNotFound, "not_found", "class not found")
}
