package model

import "time"

// RoleUser owns a school workspace. RoleAdmin only signs in to the back office
// and never owns school data.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID, Name, Email, PasswordHash string
	Role                          string
	CreatedAt                     time.Time
	BlockedAt                     *time.Time
}

func (u User) IsAdmin() bool { return u.Role == RoleAdmin }

func (u User) IsBlocked() bool { return u.BlockedAt != nil }

type Session struct {
	ID, UserID string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

type School struct {
	ID, Name string
	Grades   []string
	IsActive bool
}

type AcademicSegment struct {
	ID, Name, Type, StartDate, EndDate string
	DurationDays                       int
}

type AcademicYear struct {
	ID, SchoolID, Name, StartDate, EndDate string
	DurationDays                           int
	IsCurrent                              bool
	Segments                               []AcademicSegment
}

type Reference struct{ ID, Name string }

type StudentGrade struct {
	AcademicYearID, AcademicYearName, Grade string
	IsCurrent                               bool
}

type Student struct {
	ID, SchoolID, FirstName, LastName, Email, Phone string
	GuardianName, GuardianEmail, GuardianPhone      string
	ResidentAddress, PermanentAddress               string
	Grades                                          []StudentGrade
	Classes                                         []Reference
	Logs                                            []StudentLog
	Assignments                                     []Result
	Exams                                           []Result
}

// GradeFor returns the grade recorded for the given academic year. When the
// student has no entry for that year it falls back to the current year so
// listings still show a meaningful grade.
func (s Student) GradeFor(academicYearID string) string {
	if academicYearID != "" {
		for _, grade := range s.Grades {
			if grade.AcademicYearID == academicYearID {
				return grade.Grade
			}
		}
	}
	for _, grade := range s.Grades {
		if grade.IsCurrent {
			return grade.Grade
		}
	}
	return ""
}

func (s Student) FullName() string {
	if s.LastName == "" {
		return s.FirstName
	}
	return s.FirstName + " " + s.LastName
}

type StudentLog struct {
	ID, StudentID, Kind, Type, Note string
	CreatedAt                       time.Time
}

type Result struct {
	ID, Name, Kind, DueAt string
	Score                 *float64
	TotalScore            float64
	CompletedAt           *time.Time
}

type Class struct {
	ID, SchoolID, AcademicYearID, Name, Subject, Grade, Description string
	Students                                                        []Student
	Assignments                                                     []Reference
	Exams                                                           []Reference
}

type AssignmentStudent struct {
	Student     Student
	Score       *float64
	CompletedAt *time.Time
}

type Assignment struct {
	ID, SchoolID, AcademicYearID, Name, Type, DueDate string
	TotalScore                                        float64
	ClassID                                           *string
	Class                                             *Reference
	Students                                          []AssignmentStudent
}

type ExamStudent struct {
	Student  Student
	Score    *float64
	MarkedAt *time.Time
}

type Exam struct {
	ID, SchoolID, AcademicYearID, ClassID, Name, Type, ExamDate string
	TotalScore                                                  float64
	Class                                                       Reference
	Students                                                    []ExamStudent
}

type Overview struct {
	Students, Classes, Assignments, Exams int
	School, AcademicYear                  Reference
}
