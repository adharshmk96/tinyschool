package dataio

import (
	"fmt"
	"strings"

	"tinyschool-api/internal/model"
)

// Sheet names double as CSV file names inside an exported archive.
const (
	SheetSchools           = "schools"
	SheetAcademicYears     = "academic_years"
	SheetAcademicSegments  = "academic_segments"
	SheetStudents          = "students"
	SheetStudentClassrooms = "student_classrooms"
	SheetStudentLogs       = "student_logs"
	SheetClasses           = "classes"
	SheetClassStudents     = "class_students"
	SheetAssignments       = "assignments"
	SheetAssignmentScores  = "assignment_scores"
	SheetExams             = "exams"
	SheetExamScores        = "exam_scores"
)

// SheetNames lists every sheet in the order it is written. Importers may supply
// a subset; anything missing is treated as empty.
var SheetNames = []string{
	SheetSchools, SheetAcademicYears, SheetAcademicSegments,
	SheetStudents, SheetStudentClassrooms, SheetStudentLogs,
	SheetClasses, SheetClassStudents,
	SheetAssignments, SheetAssignmentScores,
	SheetExams, SheetExamScores,
}

var sheetColumns = map[string][]string{
	SheetSchools:           {"id", "name", "classrooms", "isActive"},
	SheetAcademicYears:     {"id", "schoolId", "name", "startDate", "endDate", "isCurrent"},
	SheetAcademicSegments:  {"id", "academicYearId", "name", "type", "startDate", "endDate"},
	SheetStudents:          {"id", "schoolId", "firstName", "lastName", "email", "phone", "guardianName", "guardianEmail", "guardianPhone", "residentAddress", "permanentAddress"},
	SheetStudentClassrooms: {"studentId", "academicYearId", "classroom"},
	SheetStudentLogs:       {"id", "studentId", "kind", "type", "note", "createdAt"},
	SheetClasses:           {"id", "schoolId", "academicYearId", "name", "subject", "classrooms", "description"},
	SheetClassStudents:     {"classId", "studentId"},
	SheetAssignments:       {"id", "schoolId", "academicYearId", "classId", "name", "type", "dueDate", "totalScore"},
	SheetAssignmentScores:  {"assignmentId", "studentId", "score", "completedAt"},
	SheetExams:             {"id", "schoolId", "academicYearId", "classId", "name", "type", "examDate", "totalScore"},
	SheetExamScores:        {"examId", "studentId", "score", "markedAt"},
}

// Columns returns the header row for a sheet, or nil when the name is unknown.
func Columns(sheet string) []string { return sheetColumns[sheet] }

// tables flattens a dataset into the sheets that get written out. Every sheet
// is present even when empty, so an export of a fresh account is also a usable
// import template.
func tables(dataset model.Dataset) []Table {
	rows := map[string][][]string{
		SheetSchools:           make([][]string, 0, len(dataset.Schools)),
		SheetAcademicYears:     make([][]string, 0, len(dataset.AcademicYears)),
		SheetAcademicSegments:  make([][]string, 0, len(dataset.AcademicSegments)),
		SheetStudents:          make([][]string, 0, len(dataset.Students)),
		SheetStudentClassrooms: make([][]string, 0, len(dataset.StudentClassrooms)),
		SheetStudentLogs:       make([][]string, 0, len(dataset.StudentLogs)),
		SheetClasses:           make([][]string, 0, len(dataset.Classes)),
		SheetClassStudents:     make([][]string, 0, len(dataset.ClassStudents)),
		SheetAssignments:       make([][]string, 0, len(dataset.Assignments)),
		SheetAssignmentScores:  make([][]string, 0, len(dataset.AssignmentScores)),
		SheetExams:             make([][]string, 0, len(dataset.Exams)),
		SheetExamScores:        make([][]string, 0, len(dataset.ExamScores)),
	}

	for _, item := range dataset.Schools {
		rows[SheetSchools] = append(rows[SheetSchools], []string{
			item.ID, item.Name, joinValues(item.Classrooms), boolCell(item.IsActive),
		})
	}
	for _, item := range dataset.AcademicYears {
		rows[SheetAcademicYears] = append(rows[SheetAcademicYears], []string{
			item.ID, item.SchoolID, item.Name, item.StartDate, item.EndDate, boolCell(item.IsCurrent),
		})
	}
	for _, item := range dataset.AcademicSegments {
		rows[SheetAcademicSegments] = append(rows[SheetAcademicSegments], []string{
			item.ID, item.AcademicYearID, item.Name, item.Type, item.StartDate, item.EndDate,
		})
	}
	for _, item := range dataset.Students {
		rows[SheetStudents] = append(rows[SheetStudents], []string{
			item.ID, item.SchoolID, item.FirstName, item.LastName, item.Email, item.Phone,
			item.GuardianName, item.GuardianEmail, item.GuardianPhone,
			item.ResidentAddress, item.PermanentAddress,
		})
	}
	for _, item := range dataset.StudentClassrooms {
		rows[SheetStudentClassrooms] = append(rows[SheetStudentClassrooms], []string{
			item.StudentID, item.AcademicYearID, item.Classroom,
		})
	}
	for _, item := range dataset.StudentLogs {
		rows[SheetStudentLogs] = append(rows[SheetStudentLogs], []string{
			item.ID, item.StudentID, item.Kind, item.Type, item.Note, timeCell(item.CreatedAt),
		})
	}
	for _, item := range dataset.Classes {
		rows[SheetClasses] = append(rows[SheetClasses], []string{
			item.ID, item.SchoolID, item.AcademicYearID, item.Name, item.Subject,
			joinValues(item.Classrooms), item.Description,
		})
	}
	for _, item := range dataset.ClassStudents {
		rows[SheetClassStudents] = append(rows[SheetClassStudents], []string{item.ClassID, item.StudentID})
	}
	for _, item := range dataset.Assignments {
		rows[SheetAssignments] = append(rows[SheetAssignments], []string{
			item.ID, item.SchoolID, item.AcademicYearID, item.ClassID, item.Name, item.Type,
			item.DueDate, numberCell(item.TotalScore),
		})
	}
	for _, item := range dataset.AssignmentScores {
		rows[SheetAssignmentScores] = append(rows[SheetAssignmentScores], []string{
			item.ParentID, item.StudentID, floatCell(item.Score), timePointerCell(item.RecordedAt),
		})
	}
	for _, item := range dataset.Exams {
		rows[SheetExams] = append(rows[SheetExams], []string{
			item.ID, item.SchoolID, item.AcademicYearID, item.ClassID, item.Name, item.Type,
			item.ExamDate, numberCell(item.TotalScore),
		})
	}
	for _, item := range dataset.ExamScores {
		rows[SheetExamScores] = append(rows[SheetExamScores], []string{
			item.ParentID, item.StudentID, floatCell(item.Score), timePointerCell(item.RecordedAt),
		})
	}

	result := make([]Table, 0, len(SheetNames))
	for _, name := range SheetNames {
		result = append(result, Table{Name: name, Columns: sheetColumns[name], Rows: rows[name]})
	}
	return result
}

// dataset rebuilds a dataset from the tables found in an uploaded file. It
// reports every cell-level problem it can see at once instead of stopping at
// the first one, so a user fixes their spreadsheet in a single pass.
func dataset(parsed []Table) (model.Dataset, error) {
	byName := make(map[string]Table, len(parsed))
	recognised := 0
	for _, table := range parsed {
		name := normaliseSheet(table.Name)
		if _, known := sheetColumns[name]; !known {
			continue
		}
		table.Name = name
		byName[name] = table
		recognised++
	}
	if recognised == 0 {
		return model.Dataset{}, fmt.Errorf("no known sheets found; expected at least one of: %s", strings.Join(SheetNames, ", "))
	}

	var errs []string
	var result model.Dataset

	forEachRow(byName[SheetSchools], &errs, func(row rowReader) {
		result.Schools = append(result.Schools, model.SchoolRow{
			ID: row.text("id"), Name: row.text("name"),
			Classrooms: row.list("classrooms"), IsActive: row.boolean("isActive"),
		})
	})
	forEachRow(byName[SheetAcademicYears], &errs, func(row rowReader) {
		result.AcademicYears = append(result.AcademicYears, model.AcademicYearRow{
			ID: row.text("id"), SchoolID: row.text("schoolId"), Name: row.text("name"),
			StartDate: row.text("startDate"), EndDate: row.text("endDate"), IsCurrent: row.boolean("isCurrent"),
		})
	})
	// Segments are ordered by their position within a year, which the file
	// expresses as row order rather than as a column.
	positions := map[string]int{}
	forEachRow(byName[SheetAcademicSegments], &errs, func(row rowReader) {
		yearID := row.text("academicYearId")
		result.AcademicSegments = append(result.AcademicSegments, model.AcademicSegmentRow{
			ID: row.text("id"), AcademicYearID: yearID, Name: row.text("name"),
			Type: strings.ToLower(row.text("type")), StartDate: row.text("startDate"), EndDate: row.text("endDate"),
			Position: positions[yearID],
		})
		positions[yearID]++
	})
	forEachRow(byName[SheetStudents], &errs, func(row rowReader) {
		result.Students = append(result.Students, model.StudentRow{
			ID: row.text("id"), SchoolID: row.text("schoolId"),
			FirstName: row.text("firstName"), LastName: row.text("lastName"),
			Email: strings.ToLower(row.text("email")), Phone: row.text("phone"),
			GuardianName: row.text("guardianName"), GuardianEmail: strings.ToLower(row.text("guardianEmail")),
			GuardianPhone:   row.text("guardianPhone"),
			ResidentAddress: row.text("residentAddress"), PermanentAddress: row.text("permanentAddress"),
		})
	})
	forEachRow(byName[SheetStudentClassrooms], &errs, func(row rowReader) {
		result.StudentClassrooms = append(result.StudentClassrooms, model.StudentClassroomRow{
			StudentID: row.text("studentId"), AcademicYearID: row.text("academicYearId"), Classroom: row.text("classroom"),
		})
	})
	forEachRow(byName[SheetStudentLogs], &errs, func(row rowReader) {
		log := model.StudentLogRow{
			ID: row.text("id"), StudentID: row.text("studentId"),
			Kind: strings.ToLower(row.text("kind")), Type: strings.ToLower(row.text("type")), Note: row.text("note"),
		}
		if created := row.timestamp("createdAt"); created != nil {
			log.CreatedAt = *created
		}
		result.StudentLogs = append(result.StudentLogs, log)
	})
	forEachRow(byName[SheetClasses], &errs, func(row rowReader) {
		result.Classes = append(result.Classes, model.ClassRow{
			ID: row.text("id"), SchoolID: row.text("schoolId"), AcademicYearID: row.text("academicYearId"),
			Name: row.text("name"), Subject: row.text("subject"),
			Classrooms: row.list("classrooms"), Description: row.text("description"),
		})
	})
	forEachRow(byName[SheetClassStudents], &errs, func(row rowReader) {
		result.ClassStudents = append(result.ClassStudents, model.ClassStudentRow{
			ClassID: row.text("classId"), StudentID: row.text("studentId"),
		})
	})
	forEachRow(byName[SheetAssignments], &errs, func(row rowReader) {
		result.Assignments = append(result.Assignments, model.AssignmentRow{
			ID: row.text("id"), SchoolID: row.text("schoolId"), AcademicYearID: row.text("academicYearId"),
			ClassID: row.text("classId"), Name: row.text("name"), Type: strings.ToLower(row.text("type")),
			DueDate: row.text("dueDate"), TotalScore: row.number("totalScore"),
		})
	})
	forEachRow(byName[SheetAssignmentScores], &errs, func(row rowReader) {
		result.AssignmentScores = append(result.AssignmentScores, model.ScoreRow{
			ParentID: row.text("assignmentId"), StudentID: row.text("studentId"),
			Score: row.optionalNumber("score"), RecordedAt: row.timestamp("completedAt"),
		})
	})
	forEachRow(byName[SheetExams], &errs, func(row rowReader) {
		result.Exams = append(result.Exams, model.ExamRow{
			ID: row.text("id"), SchoolID: row.text("schoolId"), AcademicYearID: row.text("academicYearId"),
			ClassID: row.text("classId"), Name: row.text("name"), Type: strings.ToLower(row.text("type")),
			ExamDate: row.text("examDate"), TotalScore: row.number("totalScore"),
		})
	})
	forEachRow(byName[SheetExamScores], &errs, func(row rowReader) {
		result.ExamScores = append(result.ExamScores, model.ScoreRow{
			ParentID: row.text("examId"), StudentID: row.text("studentId"),
			Score: row.optionalNumber("score"), RecordedAt: row.timestamp("markedAt"),
		})
	})

	if len(errs) > 0 {
		return model.Dataset{}, fmt.Errorf("%s", strings.Join(limitErrors(errs), "; "))
	}
	return result, nil
}

// normaliseSheet maps the file names and display variants a user might produce
// ("Students", "students.csv", "student classrooms") onto a known sheet.
func normaliseSheet(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimSuffix(name, ".csv")
	name = strings.TrimSuffix(name, ".tsv")
	if index := strings.LastIndex(name, "/"); index >= 0 {
		name = name[index+1:]
	}
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

const maxReportedErrors = 20

func limitErrors(errs []string) []string {
	if len(errs) <= maxReportedErrors {
		return errs
	}
	return append(errs[:maxReportedErrors:maxReportedErrors],
		fmt.Sprintf("and %d more problems", len(errs)-maxReportedErrors))
}
