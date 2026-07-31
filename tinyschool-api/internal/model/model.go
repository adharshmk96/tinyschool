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

// PasswordResetToken stores only the hash of the token that was handed out, so
// a leaked database cannot be used to reset anybody's password.
type PasswordResetToken struct {
	ID, UserID, TokenHash string
	ExpiresAt             time.Time
	UsedAt                *time.Time
	CreatedAt             time.Time
}

type School struct {
	ID, Name   string
	Classrooms []string
	IsActive   bool
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

type StudentClassroom struct {
	AcademicYearID, AcademicYearName, Classroom string
	IsCurrent                                   bool
}

type Student struct {
	ID, SchoolID, FirstName, LastName, Email, Phone string
	GuardianName, GuardianEmail, GuardianPhone      string
	ResidentAddress, PermanentAddress               string
	Classrooms                                      []StudentClassroom
	Classes                                         []Reference
	Logs                                            []StudentLog
	Assignments                                     []Result
	Exams                                           []Result
}

// ClassroomFor returns the classroom recorded for the given academic year. When
// the student has no entry for that year it falls back to the current year so
// listings still show a meaningful classroom.
func (s Student) ClassroomFor(academicYearID string) string {
	if academicYearID != "" {
		for _, classroom := range s.Classrooms {
			if classroom.AcademicYearID == academicYearID {
				return classroom.Classroom
			}
		}
	}
	for _, classroom := range s.Classrooms {
		if classroom.IsCurrent {
			return classroom.Classroom
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
	ID, SchoolID, AcademicYearID, Name, Subject, Description string
	Classrooms                                               []string
	Students                                                 []Student
	Assignments                                              []Reference
	Exams                                                    []Reference
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

type UpcomingItem struct {
	ID, Kind, Name, Date string
	Class                *Reference
	StudentCount         int
}
