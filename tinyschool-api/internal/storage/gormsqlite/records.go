package gormsqlite

import "time"

func (userRecord) TableName() string              { return "users" }
func (sessionRecord) TableName() string           { return "sessions" }
func (schoolRecord) TableName() string            { return "schools" }
func (schoolGradeRecord) TableName() string       { return "school_grades" }
func (academicYearRecord) TableName() string      { return "academic_years" }
func (academicSegmentRecord) TableName() string   { return "academic_segments" }
func (classRecord) TableName() string             { return "classes" }
func (classStudentRecord) TableName() string      { return "class_students" }
func (studentRecord) TableName() string           { return "students" }
func (studentGradeRecord) TableName() string      { return "student_grades" }
func (studentLogRecord) TableName() string        { return "student_logs" }
func (assignmentRecord) TableName() string        { return "assignments" }
func (assignmentStudentRecord) TableName() string { return "assignment_students" }
func (examRecord) TableName() string              { return "exams" }
func (examStudentRecord) TableName() string       { return "exam_students" }

type userRecord struct {
	ID           string `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	Email        string `gorm:"not null;uniqueIndex;collate:nocase"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:'user';index"`
	CreatedAt    time.Time
	BlockedAt    *time.Time
}

type sessionRecord struct {
	ID        string     `gorm:"primaryKey"`
	UserID    string     `gorm:"not null;index"`
	User      userRecord `gorm:"foreignKey:UserID;references:ID"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	RevokedAt *time.Time
	CreatedAt time.Time `gorm:"not null"`
}

type schoolRecord struct {
	ID       string              `gorm:"primaryKey"`
	UserID   string              `gorm:"not null;default:'';index;uniqueIndex:idx_schools_user_name,priority:1"`
	Name     string              `gorm:"not null;collate:nocase;uniqueIndex:idx_schools_user_name,priority:2"`
	IsActive bool                `gorm:"not null"`
	Grades   []schoolGradeRecord `gorm:"foreignKey:SchoolID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type schoolGradeRecord struct {
	SchoolID string `gorm:"primaryKey"`
	Grade    string `gorm:"primaryKey;collate:nocase"`
	Position int    `gorm:"not null"`
}

type academicYearRecord struct {
	ID           string                  `gorm:"primaryKey"`
	UserID       string                  `gorm:"not null;default:'';index"`
	SchoolID     string                  `gorm:"not null;index"`
	School       schoolRecord            `gorm:"foreignKey:SchoolID;references:ID"`
	Name         string                  `gorm:"not null"`
	StartDate    string                  `gorm:"not null;index"`
	EndDate      string                  `gorm:"not null"`
	DurationDays int                     `gorm:"not null"`
	IsCurrent    bool                    `gorm:"not null;index"`
	Segments     []academicSegmentRecord `gorm:"foreignKey:AcademicYearID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type academicSegmentRecord struct {
	ID             string `gorm:"primaryKey"`
	AcademicYearID string `gorm:"not null;index"`
	Name           string `gorm:"not null"`
	Type           string `gorm:"not null"`
	DurationDays   int    `gorm:"not null"`
	StartDate      string `gorm:"not null"`
	EndDate        string `gorm:"not null"`
	Position       int    `gorm:"not null"`
}

type studentRecord struct {
	ID                                         string       `gorm:"primaryKey"`
	UserID                                     string       `gorm:"not null;default:'';index"`
	SchoolID                                   string       `gorm:"not null;index"`
	School                                     schoolRecord `gorm:"foreignKey:SchoolID;references:ID"`
	FirstName, LastName, Email, Phone          string
	GuardianName, GuardianEmail, GuardianPhone string
	ResidentAddress, PermanentAddress          string
	Grades                                     []studentGradeRecord `gorm:"foreignKey:StudentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// studentGradeRecord stores the grade a student sits in for one academic year.
// Students themselves are never scoped to a year; only their grade is.
type studentGradeRecord struct {
	StudentID      string             `gorm:"primaryKey"`
	AcademicYearID string             `gorm:"primaryKey"`
	Grade          string             `gorm:"not null;collate:nocase"`
	AcademicYear   academicYearRecord `gorm:"foreignKey:AcademicYearID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type classRecord struct {
	ID                                string `gorm:"primaryKey"`
	UserID                            string `gorm:"not null;default:'';index"`
	SchoolID, AcademicYearID          string `gorm:"not null;index"`
	Name, Subject, Grade, Description string
	School                            schoolRecord       `gorm:"foreignKey:SchoolID;references:ID"`
	AcademicYear                      academicYearRecord `gorm:"foreignKey:AcademicYearID;references:ID"`
	Students                          []studentRecord    `gorm:"many2many:class_students;foreignKey:ID;joinForeignKey:ClassID;references:ID;joinReferences:StudentID"`
}

type classStudentRecord struct {
	ClassID, StudentID string `gorm:"primaryKey"`
}

type studentLogRecord struct {
	ID               string `gorm:"primaryKey"`
	StudentID        string `gorm:"not null;index"`
	Kind, Type, Note string
	CreatedAt        time.Time     `gorm:"not null;index"`
	Student          studentRecord `gorm:"foreignKey:StudentID;references:ID"`
}

type assignmentRecord struct {
	ID                       string `gorm:"primaryKey"`
	UserID                   string `gorm:"not null;default:'';index"`
	SchoolID, AcademicYearID string `gorm:"not null;index"`
	Name, Type, DueDate      string
	TotalScore               float64
	ClassID                  *string
	School                   schoolRecord              `gorm:"foreignKey:SchoolID;references:ID"`
	AcademicYear             academicYearRecord        `gorm:"foreignKey:AcademicYearID;references:ID"`
	Class                    *classRecord              `gorm:"foreignKey:ClassID;references:ID"`
	Students                 []assignmentStudentRecord `gorm:"foreignKey:AssignmentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type assignmentStudentRecord struct {
	AssignmentID, StudentID string        `gorm:"primaryKey"`
	Student                 studentRecord `gorm:"foreignKey:StudentID;references:ID"`
	Score                   *float64
	CompletedAt             *time.Time
}

type examRecord struct {
	ID                                string `gorm:"primaryKey"`
	UserID                            string `gorm:"not null;default:'';index"`
	SchoolID, AcademicYearID, ClassID string `gorm:"not null;index"`
	Name, Type, ExamDate              string
	TotalScore                        float64
	School                            schoolRecord        `gorm:"foreignKey:SchoolID;references:ID"`
	AcademicYear                      academicYearRecord  `gorm:"foreignKey:AcademicYearID;references:ID"`
	Class                             classRecord         `gorm:"foreignKey:ClassID;references:ID"`
	Students                          []examStudentRecord `gorm:"foreignKey:ExamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type examStudentRecord struct {
	ExamID, StudentID string        `gorm:"primaryKey"`
	Student           studentRecord `gorm:"foreignKey:StudentID;references:ID"`
	Score             *float64
	MarkedAt          *time.Time
}
