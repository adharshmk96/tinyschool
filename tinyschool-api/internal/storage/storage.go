package storage

import (
	"context"
	"errors"
	"time"

	"tinyschool-api/internal/model"
)

var (
	ErrNotFound = errors.New("storage: not found")
	ErrConflict = errors.New("storage: conflict")
)

type ListOptions struct {
	Search, Sort, Order string
	AcademicYearID      string
	Page, PageSize      int
}

type Storage interface {
	Ping(context.Context) error
	Close() error
	AutoMigrate(context.Context) error
	Seed(context.Context) error

	CurrentUser(context.Context) (model.User, error)
	UserByID(context.Context, string) (model.User, error)
	UserByEmail(context.Context, string) (model.User, error)
	CreateUser(context.Context, model.User, *model.School) (model.User, error)
	UpdateUser(context.Context, model.User) (model.User, error)
	CreateSession(context.Context, model.Session) (model.Session, error)
	Session(context.Context, string) (model.Session, error)
	RevokeSession(context.Context, string, time.Time) error
	RevokeOtherSessions(ctx context.Context, userID, exceptSessionID string, revokedAt time.Time) error
	ClearUserData(context.Context) error
	Overview(context.Context) (model.Overview, error)

	ListSchools(context.Context, ListOptions) ([]model.School, int64, error)
	School(context.Context, string) (model.School, error)
	CreateSchool(context.Context, model.School) (model.School, error)
	UpdateSchool(context.Context, model.School) (model.School, error)
	DeleteSchool(context.Context, string) error

	ListAcademicYears(context.Context, ListOptions) ([]model.AcademicYear, int64, error)
	AcademicYear(context.Context, string) (model.AcademicYear, error)
	CreateAcademicYear(context.Context, model.AcademicYear) (model.AcademicYear, error)
	UpdateAcademicYear(context.Context, model.AcademicYear) (model.AcademicYear, error)
	DeleteAcademicYear(context.Context, string) error

	ListClasses(context.Context, ListOptions) ([]model.Class, int64, error)
	Class(context.Context, string) (model.Class, error)
	CreateClass(context.Context, model.Class, []string) (model.Class, error)
	UpdateClass(context.Context, model.Class, *[]string) (model.Class, error)
	DeleteClass(context.Context, string) error

	ListStudents(context.Context, ListOptions) ([]model.Student, int64, error)
	Student(context.Context, string) (model.Student, error)
	CreateStudent(context.Context, model.Student) (model.Student, error)
	UpdateStudent(context.Context, model.Student) (model.Student, error)
	DeleteStudent(context.Context, string) error
	CreateStudentLog(context.Context, model.StudentLog) (model.StudentLog, error)
	DeleteStudentLog(ctx context.Context, studentID, logID, kind string) error

	ListAssignments(context.Context, ListOptions) ([]model.Assignment, int64, error)
	Assignment(context.Context, string) (model.Assignment, error)
	CreateAssignment(context.Context, model.Assignment, []string) (model.Assignment, error)
	UpdateAssignment(context.Context, model.Assignment, *[]string) (model.Assignment, error)
	DeleteAssignment(context.Context, string) error
	SetAssignmentScore(ctx context.Context, assignmentID, studentID string, score *float64, completedAt *time.Time) error

	ListExams(context.Context, ListOptions) ([]model.Exam, int64, error)
	Exam(context.Context, string) (model.Exam, error)
	CreateExam(context.Context, model.Exam) (model.Exam, error)
	UpdateExam(context.Context, model.Exam) (model.Exam, error)
	DeleteExam(context.Context, string) error
	SetExamScore(ctx context.Context, examID, studentID string, score *float64, markedAt *time.Time) error
}
