package assignments

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

type Assignee struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Grade       string   `json:"grade"`
	Score       *float64 `json:"score"`
	CompletedAt string   `json:"completedAt,omitempty"`
}

type Assignment struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	DueDate         string     `json:"dueDate"`
	TotalScore      float64    `json:"totalScore"`
	Class           *Reference `json:"class,omitempty"`
	AssigneeCount   int        `json:"assigneeCount"`
	CompletionCount int        `json:"completionCount"`
	Completion      int        `json:"completion"`
	Assignees       []Assignee `json:"assignees,omitempty"`
}

func score(value float64) *float64 { return &value }

var fixtures = []Assignment{
	{
		ID: "asg_001", Name: "Algebra Practice", Type: "class", DueDate: "2026-08-02", TotalScore: 20,
		Class: &Reference{ID: "cls_math7", Name: "Mathematics 7A"}, AssigneeCount: 3, CompletionCount: 2, Completion: 67,
		Assignees: []Assignee{
			{ID: "stu_001", Name: "Maya Patel", Grade: "Grade 7", Score: score(18), CompletedAt: "2026-07-27T09:10:00Z"},
			{ID: "stu_002", Name: "Noah Williams", Grade: "Grade 7"},
			{ID: "stu_003", Name: "Aarav Shah", Grade: "Grade 7", Score: score(16), CompletedAt: "2026-07-28T08:20:00Z"},
		},
	},
	{
		ID: "asg_002", Name: "Plant Cell Model", Type: "class", DueDate: "2026-08-08", TotalScore: 30,
		Class: &Reference{ID: "cls_sci7", Name: "Science 7A"}, AssigneeCount: 3, CompletionCount: 1, Completion: 33,
		Assignees: []Assignee{{ID: "stu_001", Name: "Maya Patel", Grade: "Grade 7"}, {ID: "stu_002", Name: "Noah Williams", Grade: "Grade 7"}, {ID: "stu_003", Name: "Aarav Shah", Grade: "Grade 7", Score: score(27), CompletedAt: "2026-07-26T13:00:00Z"}},
	},
	{
		ID: "asg_003", Name: "Short Story Draft", Type: "class", DueDate: "2026-08-10", TotalScore: 25,
		Class: &Reference{ID: "cls_eng6", Name: "English 6B"}, AssigneeCount: 2, CompletionCount: 0, Completion: 0,
		Assignees: []Assignee{{ID: "stu_004", Name: "Emma Chen", Grade: "Grade 6"}, {ID: "stu_005", Name: "Liam Brown", Grade: "Grade 6"}},
	},
	{ID: "asg_004", Name: "Geometry Project", Type: "individual", DueDate: "2026-08-18", TotalScore: 50, AssigneeCount: 1, CompletionCount: 0, Completion: 0, Assignees: []Assignee{{ID: "stu_001", Name: "Maya Patel", Grade: "Grade 7"}}},
	{
		ID: "asg_005", Name: "Ancient Civilizations Essay", Type: "class", DueDate: "2026-08-22", TotalScore: 40,
		Class: &Reference{ID: "cls_hist8", Name: "History 8A"}, AssigneeCount: 3, CompletionCount: 2, Completion: 67,
		Assignees: []Assignee{
			{ID: "stu_006", Name: "Olivia Martin", Grade: "Grade 8", Score: score(34), CompletedAt: "2026-08-20T14:05:00Z"},
			{ID: "stu_007", Name: "Ethan Wilson", Grade: "Grade 8"},
			{ID: "stu_008", Name: "Sophia Garcia", Grade: "Grade 8", Score: score(36), CompletedAt: "2026-08-21T10:40:00Z"},
		},
	},
	{ID: "asg_006", Name: "Reading Reflection", Type: "individual", DueDate: "2026-07-30", TotalScore: 10, AssigneeCount: 1, CompletionCount: 1, Completion: 100, Assignees: []Assignee{{ID: "stu_005", Name: "Liam Brown", Grade: "Grade 6", Score: score(9), CompletedAt: "2026-07-28T15:00:00Z"}}},
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/assignments", list)
	mux.HandleFunc("GET /api/v1/assignments/{id}", detail)
}

func list(w http.ResponseWriter, r *http.Request) {
	options, err := listquery.Parse(r.URL.Query(), map[string]bool{
		"name": true, "type": true, "dueDate": true, "completion": true,
	}, "dueDate")
	if err != nil {
		webjson.Error(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	result := listquery.Apply(fixtures, options,
		func(item Assignment, search string) bool {
			className := ""
			if item.Class != nil {
				className = item.Class.Name
			}
			return listquery.Contains(search, item.Name, item.Type, className)
		},
		func(a, b Assignment, field string) bool {
			switch field {
			case "name":
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			case "type":
				return a.Type < b.Type
			case "completion":
				return a.Completion < b.Completion
			default:
				return a.DueDate < b.DueDate
			}
		},
	)
	summaries := make([]Assignment, len(result.Items))
	copy(summaries, result.Items)
	for index := range summaries {
		summaries[index].Assignees = nil
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
	webjson.Error(w, http.StatusNotFound, "not_found", "assignment not found")
}
