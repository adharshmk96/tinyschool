package students

import (
	"net/http"
	"strings"

	"tinyschool-api/internal/listquery"
	"tinyschool-api/internal/webjson"
)

type Guardian struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type Log struct {
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
}

type Result struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	DueAt       string   `json:"dueAt"`
	Score       *float64 `json:"score"`
	TotalScore  float64  `json:"totalScore"`
	CompletedAt string   `json:"completedAt,omitempty"`
}

type Performance struct {
	AverageScore   float64   `json:"averageScore"`
	ClassAverage   float64   `json:"classAverage"`
	CompletionRate int       `json:"completionRate"`
	Completed      int       `json:"completed"`
	Total          int       `json:"total"`
	Standing       string    `json:"standing"`
	Trend          []float64 `json:"trend"`
}

type Student struct {
	ID               string       `json:"id"`
	FirstName        string       `json:"firstName"`
	LastName         string       `json:"lastName"`
	FullName         string       `json:"fullName"`
	Email            string       `json:"email"`
	Phone            string       `json:"phone"`
	Grade            string       `json:"grade"`
	Guardian         Guardian     `json:"guardian"`
	ResidentAddress  string       `json:"residentAddress"`
	PermanentAddress string       `json:"permanentAddress"`
	AverageScore     float64      `json:"averageScore"`
	ClassAverage     float64      `json:"classAverage"`
	Performance      *Performance `json:"performance,omitempty"`
	Behaviour        []Log        `json:"behaviour,omitempty"`
	Notes            []Log        `json:"notes,omitempty"`
	Assignments      []Result     `json:"assignments,omitempty"`
	Exams            []Result     `json:"exams,omitempty"`
}

func score(value float64) *float64 { return &value }

var fixtures = []Student{
	{
		ID: "stu_001", FirstName: "Maya", LastName: "Patel", FullName: "Maya Patel", Email: "maya.patel@example.test",
		Phone: "+91 98765 11001", Grade: "Grade 7", Guardian: Guardian{Name: "Rina Patel", Email: "rina.patel@example.test", Phone: "+91 98765 21001"},
		ResidentAddress: "12 Lake View Road, Bengaluru", PermanentAddress: "12 Lake View Road, Bengaluru",
		AverageScore: 88.5, ClassAverage: 81.2,
		Performance: &Performance{AverageScore: 88.5, ClassAverage: 81.2, CompletionRate: 86, Completed: 18, Total: 21, Standing: "Top 10%", Trend: []float64{76, 80, 79, 84, 86, 88}},
		Behaviour:   []Log{{ID: "beh_001", Type: "positive", Note: "Helped a classmate during lab work.", CreatedAt: "2026-07-24T10:30:00Z"}},
		Notes:       []Log{{ID: "note_001", Note: "Strong improvement in algebra this month.", CreatedAt: "2026-07-25T08:15:00Z"}},
		Assignments: []Result{
			{ID: "asg_001", Name: "Algebra Practice", Kind: "assignment", DueAt: "2026-08-02", Score: score(18), TotalScore: 20, CompletedAt: "2026-07-27T09:10:00Z"},
			{ID: "asg_002", Name: "Plant Cell Model", Kind: "assignment", DueAt: "2026-08-08", TotalScore: 30},
		},
		Exams: []Result{{ID: "exam_001", Name: "Mathematics Midterm", Kind: "exam", DueAt: "2026-08-15", Score: score(86), TotalScore: 100, CompletedAt: "2026-08-15T11:00:00Z"}},
	},
	{
		ID: "stu_002", FirstName: "Noah", LastName: "Williams", FullName: "Noah Williams", Email: "noah.williams@example.test",
		Phone: "+91 98765 11002", Grade: "Grade 7", Guardian: Guardian{Name: "Sophie Williams", Email: "sophie.w@example.test", Phone: "+91 98765 21002"},
		ResidentAddress: "8 Garden Street, Bengaluru", PermanentAddress: "22 Church Lane, Mysuru",
		AverageScore: 76.0, ClassAverage: 81.2,
		Performance: &Performance{AverageScore: 76, ClassAverage: 81.2, CompletionRate: 72, Completed: 13, Total: 18, Standing: "Top 50%", Trend: []float64{72, 75, 73, 78, 76, 76}},
		Behaviour:   []Log{{ID: "beh_002", Type: "need_attention", Note: "Two assignments submitted late.", CreatedAt: "2026-07-23T09:00:00Z"}},
		Notes:       []Log{{ID: "note_002", Note: "Schedule a guardian check-in.", CreatedAt: "2026-07-24T12:00:00Z"}},
		Assignments: []Result{{ID: "asg_001", Name: "Algebra Practice", Kind: "assignment", DueAt: "2026-08-02", TotalScore: 20}},
		Exams:       []Result{{ID: "exam_001", Name: "Mathematics Midterm", Kind: "exam", DueAt: "2026-08-15", Score: score(74), TotalScore: 100, CompletedAt: "2026-08-15T11:05:00Z"}},
	},
	{ID: "stu_003", FirstName: "Aarav", LastName: "Shah", FullName: "Aarav Shah", Email: "aarav.shah@example.test", Phone: "+91 98765 11003", Grade: "Grade 7", Guardian: Guardian{Name: "Neel Shah", Email: "neel.shah@example.test", Phone: "+91 98765 21003"}, ResidentAddress: "14 Palm Avenue, Bengaluru", PermanentAddress: "14 Palm Avenue, Bengaluru", AverageScore: 82, ClassAverage: 81.2},
	{ID: "stu_004", FirstName: "Emma", LastName: "Chen", FullName: "Emma Chen", Email: "emma.chen@example.test", Phone: "+91 98765 11004", Grade: "Grade 6", Guardian: Guardian{Name: "Wei Chen", Email: "wei.chen@example.test", Phone: "+91 98765 21004"}, ResidentAddress: "3 Cedar Close, Bengaluru", PermanentAddress: "3 Cedar Close, Bengaluru", AverageScore: 91, ClassAverage: 84.1},
	{ID: "stu_005", FirstName: "Liam", LastName: "Brown", FullName: "Liam Brown", Email: "liam.brown@example.test", Phone: "+91 98765 11005", Grade: "Grade 6", Guardian: Guardian{Name: "Amelia Brown", Email: "amelia.b@example.test", Phone: "+91 98765 21005"}, ResidentAddress: "19 Station Road, Bengaluru", PermanentAddress: "41 Market Street, Pune", AverageScore: 79, ClassAverage: 84.1},
	{
		ID: "stu_006", FirstName: "Olivia", LastName: "Martin", FullName: "Olivia Martin", Email: "olivia.martin@example.test",
		Phone: "+91 98765 11006", Grade: "Grade 8", Guardian: Guardian{Name: "Lucas Martin", Email: "lucas.m@example.test", Phone: "+91 98765 21006"},
		ResidentAddress: "7 Hill Road, Bengaluru", PermanentAddress: "7 Hill Road, Bengaluru", AverageScore: 77.5, ClassAverage: 77.5,
		Performance: &Performance{AverageScore: 77.5, ClassAverage: 77.5, CompletionRate: 75, Completed: 6, Total: 8, Standing: "Top 40%", Trend: []float64{68, 72, 70, 75, 77, 78}},
		Behaviour: []Log{
			{ID: "beh_006_1", Type: "positive", Note: "Led the primary-source discussion with thoughtful questions.", CreatedAt: "2026-07-26T09:35:00Z"},
			{ID: "beh_006_2", Type: "need_attention", Note: "Needs a reminder to bring the history workbook.", CreatedAt: "2026-07-22T11:10:00Z"},
		},
		Notes: []Log{
			{ID: "note_006_1", Note: "Olivia responds well to visual timelines and source maps.", CreatedAt: "2026-07-25T13:20:00Z"},
			{ID: "note_006_2", Note: "Guardian requested a progress update after the next assessment.", CreatedAt: "2026-07-20T08:45:00Z"},
		},
		Assignments: []Result{
			{ID: "asg_005", Name: "Ancient Civilizations Essay", Kind: "assignment", DueAt: "2026-08-22", Score: score(34), TotalScore: 40, CompletedAt: "2026-08-20T14:05:00Z"},
			{ID: "asg_007", Name: "Trade Routes Map", Kind: "assignment", DueAt: "2026-08-29", TotalScore: 25},
		},
		Exams: []Result{
			{ID: "exam_004", Name: "History Source Analysis", Kind: "exam", DueAt: "2026-08-30", Score: score(38), TotalScore: 50, CompletedAt: "2026-08-30T11:25:00Z"},
			{ID: "exam_005", Name: "World History Term Exam", Kind: "exam", DueAt: "2026-09-12", TotalScore: 100},
		},
	},
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/students", list)
	mux.HandleFunc("GET /api/v1/students/{id}", detail)
}

func list(w http.ResponseWriter, r *http.Request) {
	options, err := listquery.Parse(r.URL.Query(), map[string]bool{
		"name": true, "email": true, "grade": true, "averageScore": true,
	}, "name")
	if err != nil {
		webjson.Error(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	result := listquery.Apply(fixtures, options,
		func(item Student, search string) bool {
			return listquery.Contains(search, item.FullName, item.Email, item.Phone, item.Grade, item.Guardian.Name)
		},
		func(a, b Student, field string) bool {
			switch field {
			case "email":
				return strings.ToLower(a.Email) < strings.ToLower(b.Email)
			case "grade":
				return strings.ToLower(a.Grade) < strings.ToLower(b.Grade)
			case "averageScore":
				return a.AverageScore < b.AverageScore
			default:
				return strings.ToLower(a.FullName) < strings.ToLower(b.FullName)
			}
		},
	)
	summaries := make([]Student, len(result.Items))
	copy(summaries, result.Items)
	for index := range summaries {
		summaries[index].Behaviour = nil
		summaries[index].Notes = nil
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
	webjson.Error(w, http.StatusNotFound, "not_found", "student not found")
}
