package dto

type ListOptions struct {
	Search         string
	AcademicYearID string
	Grade          string
	Sort           string
	Order          string
	Page           int
	PageSize       int
}

type Page[T any] struct {
	Items    []T
	Total    int
	Page     int
	PageSize int
}

type Reference struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Grade string `json:"grade,omitempty"`
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
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// AdminUser is the back-office view of an account. It carries the moderation
// fields the school-facing User deliberately hides.
type AdminUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Blocked   bool   `json:"blocked"`
	BlockedAt string `json:"blockedAt,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type AdminStatus struct {
	AdminExists bool `json:"adminExists"`
}

type AdminSetupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	User      User
	SessionID string
	ExpiresAt string
	Token     string
}

type Overview struct {
	Students     int            `json:"students"`
	Classes      int            `json:"classes"`
	Assignments  int            `json:"assignments"`
	Exams        int            `json:"exams"`
	School       Reference      `json:"school"`
	AcademicYear Reference      `json:"academicYear"`
	Upcoming     []UpcomingItem `json:"upcoming"`
}

type UpcomingItem struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Title        string `json:"title"`
	Date         string `json:"date"`
	ClassName    string `json:"className,omitempty"`
	StudentCount int    `json:"studentCount"`
}

type School struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Grades       []string `json:"grades"`
	GradesInUse  []string `json:"gradesInUse"`
	IsActive     bool     `json:"isActive"`
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
	AverageScore           float64            `json:"averageScore"`
	ClassAverage           float64            `json:"classAverage,omitempty"`
	CompletionRate         int                `json:"completionRate"`
	Completed              int                `json:"completed"`
	Total                  int                `json:"total"`
	Standing               string             `json:"standing"`
	Trend                  []float64          `json:"trend"`
	HighestScore           float64            `json:"highestScore,omitempty"`
	LowestScore            float64            `json:"lowestScore,omitempty"`
	MedianScore            float64            `json:"medianScore,omitempty"`
	AssignmentCompleted    int                `json:"assignmentCompleted,omitempty"`
	AssignmentTotal        int                `json:"assignmentTotal,omitempty"`
	AssignmentCompletion   int                `json:"assignmentCompletion,omitempty"`
	ExamCompleted          int                `json:"examCompleted,omitempty"`
	ExamTotal              int                `json:"examTotal,omitempty"`
	ExamCompletion         int                `json:"examCompletion,omitempty"`
	AssignmentAverage      float64            `json:"assignmentAverage,omitempty"`
	ExamAverage            float64            `json:"examAverage,omitempty"`
	StrongestArea          string             `json:"strongestArea,omitempty"`
	ScoreConsistency       int                `json:"scoreConsistency,omitempty"`
	BestResult             *PerformanceResult `json:"bestResult,omitempty"`
	TopStudents            []RankedStudent    `json:"topStudents,omitempty"`
	MonthlyExamTrend       []TrendPoint       `json:"monthlyExamTrend,omitempty"`
	MonthlyAssignmentTrend []TrendPoint       `json:"monthlyAssignmentTrend,omitempty"`
}

type PerformanceResult struct {
	Name  string  `json:"name"`
	Kind  string  `json:"kind"`
	Score float64 `json:"score"`
}

type RankedStudent struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

type TrendPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
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

// StudentGradeInput links one academic year to the grade the student sits in.
type StudentGradeInput struct {
	AcademicYearID string `json:"academicYearId"`
	Grade          string `json:"grade"`
}

type StudentGrade struct {
	AcademicYearID   string `json:"academicYearId"`
	AcademicYearName string `json:"academicYearName"`
	Grade            string `json:"grade"`
	IsCurrent        bool   `json:"isCurrent"`
}

type StudentRequest struct {
	SchoolID         string              `json:"schoolId"`
	FirstName        string              `json:"firstName"`
	LastName         string              `json:"lastName"`
	Email            string              `json:"email"`
	Phone            string              `json:"phone"`
	Grades           []StudentGradeInput `json:"grades"`
	Guardian         Guardian            `json:"guardian"`
	ResidentAddress  string              `json:"residentAddress"`
	PermanentAddress string              `json:"permanentAddress"`
}

type UpdateStudentRequest struct {
	SchoolID         *string              `json:"schoolId"`
	FirstName        *string              `json:"firstName"`
	LastName         *string              `json:"lastName"`
	Email            *string              `json:"email"`
	Phone            *string              `json:"phone"`
	Grades           *[]StudentGradeInput `json:"grades"`
	Guardian         *Guardian            `json:"guardian"`
	ResidentAddress  *string              `json:"residentAddress"`
	PermanentAddress *string              `json:"permanentAddress"`
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
	Grades           []StudentGrade  `json:"grades"`
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
	Performance     *Performance         `json:"performance,omitempty"`
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
