package model

import "time"

// Dataset is the complete workspace owned by one account, flattened into the
// rows a spreadsheet can hold. Every slice becomes one sheet of an exported
// workbook (or one CSV file inside the exported archive), and rows point at
// each other through the ids carried in the file rather than through nesting.
type Dataset struct {
	Schools           []SchoolRow
	AcademicYears     []AcademicYearRow
	AcademicSegments  []AcademicSegmentRow
	Students          []StudentRow
	StudentClassrooms []StudentClassroomRow
	StudentLogs       []StudentLogRow
	Classes           []ClassRow
	ClassStudents     []ClassStudentRow
	Assignments       []AssignmentRow
	AssignmentScores  []ScoreRow
	Exams             []ExamRow
	ExamScores        []ScoreRow
}

// IsEmpty reports whether the dataset carries no rows at all.
func (d Dataset) IsEmpty() bool {
	return len(d.Schools) == 0 && len(d.AcademicYears) == 0 && len(d.Students) == 0 &&
		len(d.Classes) == 0 && len(d.Assignments) == 0 && len(d.Exams) == 0
}

type SchoolRow struct {
	ID, Name   string
	Classrooms []string
	IsActive   bool
}

type AcademicYearRow struct {
	ID, SchoolID, Name, StartDate, EndDate string
	IsCurrent                              bool
}

type AcademicSegmentRow struct {
	ID, AcademicYearID, Name, Type, StartDate, EndDate string
	Position                                           int
}

type StudentRow struct {
	ID, SchoolID, FirstName, LastName, Email, Phone string
	GuardianName, GuardianEmail, GuardianPhone      string
	ResidentAddress, PermanentAddress               string
}

type StudentClassroomRow struct {
	StudentID, AcademicYearID, Classroom string
}

type StudentLogRow struct {
	ID, StudentID, Kind, Type, Note string
	CreatedAt                       time.Time
}

type ClassRow struct {
	ID, SchoolID, AcademicYearID, Name, Subject, Description string
	Classrooms                                               []string
}

type ClassStudentRow struct {
	ClassID, StudentID string
}

type AssignmentRow struct {
	ID, SchoolID, AcademicYearID, Name, Type, DueDate string
	ClassID                                           string
	TotalScore                                        float64
}

type ExamRow struct {
	ID, SchoolID, AcademicYearID, ClassID, Name, Type, ExamDate string
	TotalScore                                                  float64
}

// ScoreRow carries one student's result for an assignment or an exam. ParentID
// is the assignment or exam id, and RecordedAt is the completion or marking
// time; both sheets share the shape so they share the codec.
type ScoreRow struct {
	ParentID, StudentID string
	Score               *float64
	RecordedAt          *time.Time
}
