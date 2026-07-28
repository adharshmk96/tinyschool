package dto

type ListOptions struct {
	Search   string
	Sort     string
	Order    string
	Page     int
	PageSize int
}

type Page[T any] struct {
	Items    []T
	Total    int
	Page     int
	PageSize int
}

type Reference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type RegisterRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	SchoolName string `json:"schoolName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type AuthResult struct {
	User      User
	SessionID string
	ExpiresAt string
	Token     string
}

type Overview struct {
	Students     int       `json:"students"`
	Classes      int       `json:"classes"`
	Assignments  int       `json:"assignments"`
	Exams        int       `json:"exams"`
	School       Reference `json:"school"`
	AcademicYear Reference `json:"academicYear"`
}

type School struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Grades   []string `json:"grades"`
	IsActive bool     `json:"isActive"`
}

type SchoolRequest struct {
	Name     string   `json:"name"`
	Grades   []string `json:"grades"`
	IsActive bool     `json:"isActive"`
}

type SegmentRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DurationDays int    `json:"durationDays"`
}

type AcademicYearRequest struct {
	SchoolID  string           `json:"schoolId"`
	Name      string           `json:"name"`
	StartDate string           `json:"startDate"`
	IsCurrent bool             `json:"isCurrent"`
	Segments  []SegmentRequest `json:"segments"`
}

type AcademicSegment struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	DurationDays int    `json:"durationDays"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
}

type AcademicYear struct {
	ID           string            `json:"id"`
	SchoolID     string            `json:"schoolId"`
	Name         string            `json:"name"`
	StartDate    string            `json:"startDate"`
	EndDate      string            `json:"endDate"`
	DurationDays int               `json:"durationDays"`
	IsCurrent    bool              `json:"isCurrent"`
	Segments     []AcademicSegment `json:"segments"`
}

type ClassRequest struct {
	SchoolID       string   `json:"schoolId"`
	AcademicYearID string   `json:"academicYearId"`
	Name           string   `json:"name"`
	Subject        string   `json:"subject"`
	Grade          string   `json:"grade"`
	Description    string   `json:"description"`
	StudentIDs     []string `json:"studentIds"`
}

type UpdateClassRequest struct {
	SchoolID       *string   `json:"schoolId"`
	AcademicYearID *string   `json:"academicYearId"`
	Name           *string   `json:"name"`
	Subject        *string   `json:"subject"`
	Grade          *string   `json:"grade"`
	Description    *string   `json:"description"`
	StudentIDs     *[]string `json:"studentIds"`
}

type Performance struct {
	AverageScore   float64   `json:"averageScore"`
	ClassAverage   float64   `json:"classAverage,omitempty"`
	CompletionRate int       `json:"completionRate"`
	Completed      int       `json:"completed"`
	Total          int       `json:"total"`
	Standing       string    `json:"standing"`
	Trend          []float64 `json:"trend"`
}

type Class struct {
	ID             string       `json:"id"`
	SchoolID       string       `json:"schoolId"`
	AcademicYearID string       `json:"academicYearId"`
	Name           string       `json:"name"`
	Subject        string       `json:"subject"`
	Grade          string       `json:"grade"`
	Description    string       `json:"description"`
	StudentCount   int          `json:"studentCount"`
	AverageScore   float64      `json:"averageScore"`
	Performance    *Performance `json:"performance,omitempty"`
	Students       []Reference  `json:"students,omitempty"`
	Assignments    []Reference  `json:"assignments,omitempty"`
	Exams          []Reference  `json:"exams,omitempty"`
}

type Guardian struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type StudentRequest struct {
	SchoolID         string   `json:"schoolId"`
	FirstName        string   `json:"firstName"`
	LastName         string   `json:"lastName"`
	Email            string   `json:"email"`
	Phone            string   `json:"phone"`
	Grade            string   `json:"grade"`
	Guardian         Guardian `json:"guardian"`
	ResidentAddress  string   `json:"residentAddress"`
	PermanentAddress string   `json:"permanentAddress"`
}

type UpdateStudentRequest struct {
	SchoolID         *string   `json:"schoolId"`
	FirstName        *string   `json:"firstName"`
	LastName         *string   `json:"lastName"`
	Email            *string   `json:"email"`
	Phone            *string   `json:"phone"`
	Grade            *string   `json:"grade"`
	Guardian         *Guardian `json:"guardian"`
	ResidentAddress  *string   `json:"residentAddress"`
	PermanentAddress *string   `json:"permanentAddress"`
}

type StudentLog struct {
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
}

type StudentResult struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	DueAt       string   `json:"dueAt"`
	Score       *float64 `json:"score"`
	TotalScore  float64  `json:"totalScore"`
	CompletedAt string   `json:"completedAt,omitempty"`
}

type Student struct {
	ID               string          `json:"id"`
	SchoolID         string          `json:"schoolId"`
	FirstName        string          `json:"firstName"`
	LastName         string          `json:"lastName"`
	FullName         string          `json:"fullName"`
	Email            string          `json:"email"`
	Phone            string          `json:"phone"`
	Grade            string          `json:"grade"`
	Guardian         Guardian        `json:"guardian"`
	ResidentAddress  string          `json:"residentAddress"`
	PermanentAddress string          `json:"permanentAddress"`
	AverageScore     float64         `json:"averageScore"`
	ClassAverage     float64         `json:"classAverage"`
	Performance      *Performance    `json:"performance,omitempty"`
	Behaviour        []StudentLog    `json:"behaviour,omitempty"`
	Notes            []StudentLog    `json:"notes,omitempty"`
	Assignments      []StudentResult `json:"assignments,omitempty"`
	Exams            []StudentResult `json:"exams,omitempty"`
}

type BehaviourRequest struct {
	Type string `json:"type"`
	Note string `json:"note"`
}

type NoteRequest struct {
	Note string `json:"note"`
}

type AssignmentRequest struct {
	SchoolID       string   `json:"schoolId"`
	AcademicYearID string   `json:"academicYearId"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	DueDate        string   `json:"dueDate"`
	TotalScore     float64  `json:"totalScore"`
	ClassID        string   `json:"classId"`
	StudentIDs     []string `json:"studentIds"`
}

type UpdateAssignmentRequest struct {
	SchoolID       *string   `json:"schoolId"`
	AcademicYearID *string   `json:"academicYearId"`
	Name           *string   `json:"name"`
	Type           *string   `json:"type"`
	DueDate        *string   `json:"dueDate"`
	TotalScore     *float64  `json:"totalScore"`
	ClassID        *string   `json:"classId"`
	StudentIDs     *[]string `json:"studentIds"`
}

type AssignmentAssignee struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Grade       string   `json:"grade"`
	Score       *float64 `json:"score"`
	CompletedAt string   `json:"completedAt,omitempty"`
}

type Assignment struct {
	ID              string               `json:"id"`
	SchoolID        string               `json:"schoolId"`
	AcademicYearID  string               `json:"academicYearId"`
	Name            string               `json:"name"`
	Type            string               `json:"type"`
	DueDate         string               `json:"dueDate"`
	TotalScore      float64              `json:"totalScore"`
	Class           *Reference           `json:"class,omitempty"`
	AssigneeCount   int                  `json:"assigneeCount"`
	CompletionCount int                  `json:"completionCount"`
	Completion      int                  `json:"completion"`
	Assignees       []AssignmentAssignee `json:"assignees,omitempty"`
}

type ScoreRequest struct {
	Score float64 `json:"score"`
}

type ExamRequest struct {
	SchoolID       string  `json:"schoolId"`
	AcademicYearID string  `json:"academicYearId"`
	ClassID        string  `json:"classId"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	ExamDate       string  `json:"examDate"`
	TotalScore     float64 `json:"totalScore"`
}

type UpdateExamRequest struct {
	SchoolID       *string  `json:"schoolId"`
	AcademicYearID *string  `json:"academicYearId"`
	ClassID        *string  `json:"classId"`
	Name           *string  `json:"name"`
	Type           *string  `json:"type"`
	ExamDate       *string  `json:"examDate"`
	TotalScore     *float64 `json:"totalScore"`
}

type ExamStudent struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Grade    string   `json:"grade"`
	Score    *float64 `json:"score"`
	MarkedAt string   `json:"markedAt,omitempty"`
}

type Exam struct {
	ID             string        `json:"id"`
	SchoolID       string        `json:"schoolId"`
	AcademicYearID string        `json:"academicYearId"`
	Name           string        `json:"name"`
	Type           string        `json:"type"`
	ExamDate       string        `json:"examDate"`
	TotalScore     float64       `json:"totalScore"`
	Class          Reference     `json:"class"`
	StudentCount   int           `json:"studentCount"`
	MarkedCount    int           `json:"markedCount"`
	AverageScore   float64       `json:"averageScore"`
	Performance    *Performance  `json:"performance,omitempty"`
	Students       []ExamStudent `json:"students,omitempty"`
}
