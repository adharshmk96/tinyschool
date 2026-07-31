package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
)

// ExportData returns every row the caller owns, ready to be written to a
// spreadsheet by the delivery layer.
func (a *App) ExportData(ctx context.Context) (model.Dataset, error) {
	return a.storage.ExportUserData(ctx)
}

// ImportData validates an uploaded dataset and replaces the caller's workspace
// with it. The whole file is checked before anything is written, and the ids
// inside the file are only used to link rows to each other: every record is
// stored under a freshly generated id, so a file exported from one account can
// be imported into another without colliding.
func (a *App) ImportData(ctx context.Context, incoming model.Dataset) (dto.ImportSummary, error) {
	if incoming.IsEmpty() {
		return dto.ImportSummary{}, validation("the file contains no data to import")
	}
	prepared, err := newImport(a).build(incoming)
	if err != nil {
		return dto.ImportSummary{}, err
	}
	if err := a.storage.ReplaceUserData(ctx, prepared); err != nil {
		return dto.ImportSummary{}, translate(err, "data")
	}
	return dto.ImportSummary{
		Schools: len(prepared.Schools), AcademicYears: len(prepared.AcademicYears),
		Students: len(prepared.Students), StudentLogs: len(prepared.StudentLogs),
		Classes: len(prepared.Classes), Assignments: len(prepared.Assignments),
		Exams:  len(prepared.Exams),
		Scores: len(prepared.AssignmentScores) + len(prepared.ExamScores),
	}, nil
}

// dataImport rewrites file-local ids to storage ids while checking that every
// reference resolves. Problems are collected rather than returned one at a
// time, so a user can fix their spreadsheet in a single pass.
type dataImport struct {
	app    *App
	errs   []string
	failed bool

	schools     map[string]string
	schoolRows  map[string]int      // file school id -> index in the prepared schools
	classrooms  map[string][]string // file school id -> classrooms offered
	years       map[string]string
	yearSchool  map[string]string // file year id -> file school id
	students    map[string]string
	studentHome map[string]string // file student id -> file school id
	classes     map[string]string
	assignments map[string]string
	exams       map[string]string
}

func newImport(app *App) *dataImport {
	return &dataImport{
		app:         app,
		schools:     map[string]string{},
		schoolRows:  map[string]int{},
		classrooms:  map[string][]string{},
		years:       map[string]string{},
		yearSchool:  map[string]string{},
		students:    map[string]string{},
		studentHome: map[string]string{},
		classes:     map[string]string{},
		assignments: map[string]string{},
		exams:       map[string]string{},
	}
}

const maxImportErrors = 20

func (i *dataImport) fail(sheet string, entry int, format string, args ...any) {
	i.failed = true
	if len(i.errs) >= maxImportErrors {
		return
	}
	i.errs = append(i.errs, fmt.Sprintf("%s entry %d: %s", sheet, entry, fmt.Sprintf(format, args...)))
}

func (i *dataImport) result(dataset model.Dataset) (model.Dataset, error) {
	if !i.failed {
		return dataset, nil
	}
	message := strings.Join(i.errs, "; ")
	if len(i.errs) == maxImportErrors {
		message += "; more problems were found after the first " + fmt.Sprint(maxImportErrors)
	}
	return model.Dataset{}, validation(message)
}

// id generates the stored id for an imported row, falling back to a blank id
// only when generation fails (which also records the failure).
func (i *dataImport) id(prefix, sheet string, entry int) string {
	generated, err := i.app.newID(prefix)
	if err != nil {
		i.fail(sheet, entry, "could not generate an id")
		return ""
	}
	return generated
}

func (i *dataImport) build(incoming model.Dataset) (model.Dataset, error) {
	var out model.Dataset
	i.buildSchools(incoming, &out)
	i.buildYears(incoming, &out)
	i.buildStudents(incoming, &out)
	i.buildClasses(incoming, &out)
	i.buildAssignments(incoming, &out)
	i.buildExams(incoming, &out)
	return i.result(out)
}

func (i *dataImport) buildSchools(incoming model.Dataset, out *model.Dataset) {
	active := 0
	for index, row := range incoming.Schools {
		entry := index + 1
		key := strings.TrimSpace(row.ID)
		name := strings.TrimSpace(row.Name)
		switch {
		case key == "":
			i.fail("schools", entry, "id is required so other sheets can refer to this school")
			continue
		case name == "":
			i.fail("schools", entry, "name is required")
			continue
		}
		if _, exists := i.schools[key]; exists {
			i.fail("schools", entry, "id %q is used more than once", key)
			continue
		}
		classrooms, err := uniqueTrimmed(row.Classrooms, "classrooms")
		if err != nil {
			i.fail("schools", entry, "%s", err.Error())
			continue
		}
		if len(classrooms) == 0 {
			i.fail("schools", entry, "at least one classroom is required")
			continue
		}
		id := i.id("sch", "schools", entry)
		i.schools[key] = id
		i.schoolRows[key] = len(out.Schools)
		i.classrooms[key] = classrooms
		if row.IsActive {
			active++
		}
		out.Schools = append(out.Schools, model.SchoolRow{
			ID: id, Name: name, Classrooms: classrooms, IsActive: row.IsActive,
		})
	}
	// The workspace always opens on an active school, so pick one when the file
	// marks none.
	if active == 0 && len(out.Schools) > 0 {
		out.Schools[0].IsActive = true
	}
}

func (i *dataImport) buildYears(incoming model.Dataset, out *model.Dataset) {
	current := map[string]bool{}
	var owners []string // school key of each row appended to out.AcademicYears
	for index, row := range incoming.AcademicYears {
		entry := index + 1
		key := strings.TrimSpace(row.ID)
		name := strings.TrimSpace(row.Name)
		schoolKey := strings.TrimSpace(row.SchoolID)
		switch {
		case key == "":
			i.fail("academic_years", entry, "id is required so other sheets can refer to this year")
			continue
		case name == "":
			i.fail("academic_years", entry, "name is required")
			continue
		}
		if _, exists := i.years[key]; exists {
			i.fail("academic_years", entry, "id %q is used more than once", key)
			continue
		}
		schoolID, known := i.schools[schoolKey]
		if !known {
			i.fail("academic_years", entry, "schoolId %q is not listed in the schools sheet", schoolKey)
			continue
		}
		start, err := date(row.StartDate, "startDate")
		if err != nil {
			i.fail("academic_years", entry, "%s", err.Error())
			continue
		}
		end, err := date(row.EndDate, "endDate")
		if err != nil {
			i.fail("academic_years", entry, "%s", err.Error())
			continue
		}
		if end.Before(start) {
			i.fail("academic_years", entry, "endDate must not be before startDate")
			continue
		}
		if row.IsCurrent && current[schoolKey] {
			i.fail("academic_years", entry, "a school can only have one current academic year")
			continue
		}
		current[schoolKey] = current[schoolKey] || row.IsCurrent

		id := i.id("ay", "academic_years", entry)
		i.years[key] = id
		i.yearSchool[key] = schoolKey
		owners = append(owners, schoolKey)
		out.AcademicYears = append(out.AcademicYears, model.AcademicYearRow{
			ID: id, SchoolID: schoolID, Name: name,
			StartDate: start.Format(time.DateOnly), EndDate: end.Format(time.DateOnly),
			IsCurrent: row.IsCurrent,
		})
	}
	// Each school needs one current year for the dashboard to resolve.
	for index, schoolKey := range owners {
		if !current[schoolKey] {
			out.AcademicYears[index].IsCurrent = true
			current[schoolKey] = true
		}
	}

	for index, row := range incoming.AcademicSegments {
		entry := index + 1
		yearID, known := i.years[strings.TrimSpace(row.AcademicYearID)]
		if !known {
			i.fail("academic_segments", entry, "academicYearId %q is not listed in the academic_years sheet", row.AcademicYearID)
			continue
		}
		name := strings.TrimSpace(row.Name)
		segmentType := strings.ToLower(strings.TrimSpace(row.Type))
		if name == "" {
			i.fail("academic_segments", entry, "name is required")
			continue
		}
		if segmentType != "term" && segmentType != "vacation" {
			i.fail("academic_segments", entry, "type must be term or vacation")
			continue
		}
		start, err := date(row.StartDate, "startDate")
		if err != nil {
			i.fail("academic_segments", entry, "%s", err.Error())
			continue
		}
		end, err := date(row.EndDate, "endDate")
		if err != nil {
			i.fail("academic_segments", entry, "%s", err.Error())
			continue
		}
		if end.Before(start) {
			i.fail("academic_segments", entry, "endDate must not be before startDate")
			continue
		}
		out.AcademicSegments = append(out.AcademicSegments, model.AcademicSegmentRow{
			ID: i.id("seg", "academic_segments", entry), AcademicYearID: yearID,
			Name: name, Type: segmentType,
			StartDate: start.Format(time.DateOnly), EndDate: end.Format(time.DateOnly),
			Position: row.Position,
		})
	}
}

func (i *dataImport) buildStudents(incoming model.Dataset, out *model.Dataset) {
	for index, row := range incoming.Students {
		entry := index + 1
		key := strings.TrimSpace(row.ID)
		if key == "" {
			i.fail("students", entry, "id is required so other sheets can refer to this student")
			continue
		}
		if _, exists := i.students[key]; exists {
			i.fail("students", entry, "id %q is used more than once", key)
			continue
		}
		schoolKey := strings.TrimSpace(row.SchoolID)
		schoolID, known := i.schools[schoolKey]
		if !known {
			i.fail("students", entry, "schoolId %q is not listed in the schools sheet", schoolKey)
			continue
		}
		firstName := strings.TrimSpace(row.FirstName)
		lastName := strings.TrimSpace(row.LastName)
		if firstName == "" || lastName == "" {
			i.fail("students", entry, "firstName and lastName are required")
			continue
		}
		email := strings.ToLower(strings.TrimSpace(row.Email))
		if email != "" {
			if err := validEmail(email); err != nil {
				i.fail("students", entry, "email must be valid")
				continue
			}
		}
		guardianEmail := strings.ToLower(strings.TrimSpace(row.GuardianEmail))
		if guardianEmail != "" {
			if err := validEmail(guardianEmail); err != nil {
				i.fail("students", entry, "guardianEmail must be valid")
				continue
			}
		}
		id := i.id("stu", "students", entry)
		i.students[key] = id
		i.studentHome[key] = schoolKey
		out.Students = append(out.Students, model.StudentRow{
			ID: id, SchoolID: schoolID, FirstName: firstName, LastName: lastName,
			Email: email, Phone: strings.TrimSpace(row.Phone),
			GuardianName: strings.TrimSpace(row.GuardianName), GuardianEmail: guardianEmail,
			GuardianPhone:    strings.TrimSpace(row.GuardianPhone),
			ResidentAddress:  strings.TrimSpace(row.ResidentAddress),
			PermanentAddress: strings.TrimSpace(row.PermanentAddress),
		})
	}

	seen := map[string]bool{}
	for index, row := range incoming.StudentClassrooms {
		entry := index + 1
		studentKey := strings.TrimSpace(row.StudentID)
		studentID, known := i.students[studentKey]
		if !known {
			i.fail("student_classrooms", entry, "studentId %q is not listed in the students sheet", studentKey)
			continue
		}
		yearKey := strings.TrimSpace(row.AcademicYearID)
		yearID, known := i.years[yearKey]
		if !known {
			i.fail("student_classrooms", entry, "academicYearId %q is not listed in the academic_years sheet", yearKey)
			continue
		}
		if i.yearSchool[yearKey] != i.studentHome[studentKey] {
			i.fail("student_classrooms", entry, "the academic year belongs to a different school than the student")
			continue
		}
		classroom := strings.TrimSpace(row.Classroom)
		if classroom == "" {
			i.fail("student_classrooms", entry, "classroom is required")
			continue
		}
		i.adopt(i.studentHome[studentKey], classroom, out)
		pair := studentKey + "\x00" + yearKey
		if seen[pair] {
			i.fail("student_classrooms", entry, "the student already has a classroom for this academic year")
			continue
		}
		seen[pair] = true
		out.StudentClassrooms = append(out.StudentClassrooms, model.StudentClassroomRow{
			StudentID: studentID, AcademicYearID: yearID, Classroom: classroom,
		})
	}

	for index, row := range incoming.StudentLogs {
		entry := index + 1
		studentID, known := i.students[strings.TrimSpace(row.StudentID)]
		if !known {
			i.fail("student_logs", entry, "studentId %q is not listed in the students sheet", row.StudentID)
			continue
		}
		kind := strings.ToLower(strings.TrimSpace(row.Kind))
		if kind != "behaviour" && kind != "note" {
			i.fail("student_logs", entry, "kind must be behaviour or note")
			continue
		}
		note := strings.TrimSpace(row.Note)
		if note == "" {
			i.fail("student_logs", entry, "note is required")
			continue
		}
		createdAt := row.CreatedAt
		if createdAt.IsZero() {
			createdAt = i.app.now().UTC()
		}
		prefix := "note"
		logType := ""
		if kind == "behaviour" {
			prefix = "beh"
			logType = strings.ToLower(strings.TrimSpace(row.Type))
		}
		out.StudentLogs = append(out.StudentLogs, model.StudentLogRow{
			ID: i.id(prefix, "student_logs", entry), StudentID: studentID,
			Kind: kind, Type: logType, Note: note, CreatedAt: createdAt.UTC(),
		})
	}
}

func (i *dataImport) buildClasses(incoming model.Dataset, out *model.Dataset) {
	for index, row := range incoming.Classes {
		entry := index + 1
		key := strings.TrimSpace(row.ID)
		if key == "" {
			i.fail("classes", entry, "id is required so other sheets can refer to this class")
			continue
		}
		if _, exists := i.classes[key]; exists {
			i.fail("classes", entry, "id %q is used more than once", key)
			continue
		}
		schoolKey, yearKey := strings.TrimSpace(row.SchoolID), strings.TrimSpace(row.AcademicYearID)
		schoolID, yearID, ok := i.link("classes", entry, schoolKey, yearKey)
		if !ok {
			continue
		}
		name, subject := strings.TrimSpace(row.Name), strings.TrimSpace(row.Subject)
		if name == "" || subject == "" {
			i.fail("classes", entry, "name and subject are required")
			continue
		}
		classrooms, err := uniqueTrimmed(row.Classrooms, "classrooms")
		if err != nil {
			i.fail("classes", entry, "%s", err.Error())
			continue
		}
		if len(classrooms) == 0 {
			i.fail("classes", entry, "at least one classroom is required")
			continue
		}
		for _, classroom := range classrooms {
			i.adopt(schoolKey, classroom, out)
		}
		id := i.id("cls", "classes", entry)
		i.classes[key] = id
		out.Classes = append(out.Classes, model.ClassRow{
			ID: id, SchoolID: schoolID, AcademicYearID: yearID, Name: name, Subject: subject,
			Classrooms: classrooms, Description: strings.TrimSpace(row.Description),
		})
	}

	seen := map[string]bool{}
	for index, row := range incoming.ClassStudents {
		entry := index + 1
		classKey, studentKey := strings.TrimSpace(row.ClassID), strings.TrimSpace(row.StudentID)
		classID, known := i.classes[classKey]
		if !known {
			i.fail("class_students", entry, "classId %q is not listed in the classes sheet", classKey)
			continue
		}
		studentID, known := i.students[studentKey]
		if !known {
			i.fail("class_students", entry, "studentId %q is not listed in the students sheet", studentKey)
			continue
		}
		pair := classKey + "\x00" + studentKey
		if seen[pair] {
			continue
		}
		seen[pair] = true
		out.ClassStudents = append(out.ClassStudents, model.ClassStudentRow{ClassID: classID, StudentID: studentID})
	}
}

func (i *dataImport) buildAssignments(incoming model.Dataset, out *model.Dataset) {
	for index, row := range incoming.Assignments {
		entry := index + 1
		key := strings.TrimSpace(row.ID)
		if key == "" {
			i.fail("assignments", entry, "id is required so scores can refer to this assignment")
			continue
		}
		if _, exists := i.assignments[key]; exists {
			i.fail("assignments", entry, "id %q is used more than once", key)
			continue
		}
		schoolID, yearID, ok := i.link("assignments", entry, strings.TrimSpace(row.SchoolID), strings.TrimSpace(row.AcademicYearID))
		if !ok {
			continue
		}
		name := strings.TrimSpace(row.Name)
		if name == "" {
			i.fail("assignments", entry, "name is required")
			continue
		}
		dueDate, err := date(row.DueDate, "dueDate")
		if err != nil {
			i.fail("assignments", entry, "%s", err.Error())
			continue
		}
		if row.TotalScore <= 0 {
			i.fail("assignments", entry, "totalScore must be greater than zero")
			continue
		}
		classID := ""
		if classKey := strings.TrimSpace(row.ClassID); classKey != "" {
			resolved, known := i.classes[classKey]
			if !known {
				i.fail("assignments", entry, "classId %q is not listed in the classes sheet", classKey)
				continue
			}
			classID = resolved
		}
		id := i.id("asg", "assignments", entry)
		i.assignments[key] = id
		out.Assignments = append(out.Assignments, model.AssignmentRow{
			ID: id, SchoolID: schoolID, AcademicYearID: yearID, ClassID: classID,
			Name: name, Type: strings.ToLower(strings.TrimSpace(row.Type)),
			DueDate: dueDate.Format(time.DateOnly), TotalScore: row.TotalScore,
		})
	}
	out.AssignmentScores = i.scores("assignment_scores", "assignmentId", incoming.AssignmentScores, i.assignments)
}

func (i *dataImport) buildExams(incoming model.Dataset, out *model.Dataset) {
	for index, row := range incoming.Exams {
		entry := index + 1
		key := strings.TrimSpace(row.ID)
		if key == "" {
			i.fail("exams", entry, "id is required so scores can refer to this exam")
			continue
		}
		if _, exists := i.exams[key]; exists {
			i.fail("exams", entry, "id %q is used more than once", key)
			continue
		}
		schoolID, yearID, ok := i.link("exams", entry, strings.TrimSpace(row.SchoolID), strings.TrimSpace(row.AcademicYearID))
		if !ok {
			continue
		}
		classID, known := i.classes[strings.TrimSpace(row.ClassID)]
		if !known {
			i.fail("exams", entry, "classId %q is not listed in the classes sheet", row.ClassID)
			continue
		}
		name := strings.TrimSpace(row.Name)
		if name == "" {
			i.fail("exams", entry, "name is required")
			continue
		}
		examDate, err := date(row.ExamDate, "examDate")
		if err != nil {
			i.fail("exams", entry, "%s", err.Error())
			continue
		}
		if row.TotalScore <= 0 {
			i.fail("exams", entry, "totalScore must be greater than zero")
			continue
		}
		id := i.id("exam", "exams", entry)
		i.exams[key] = id
		out.Exams = append(out.Exams, model.ExamRow{
			ID: id, SchoolID: schoolID, AcademicYearID: yearID, ClassID: classID,
			Name: name, Type: strings.ToLower(strings.TrimSpace(row.Type)),
			ExamDate: examDate.Format(time.DateOnly), TotalScore: row.TotalScore,
		})
	}
	out.ExamScores = i.scores("exam_scores", "examId", incoming.ExamScores, i.exams)
}

// scores resolves one of the two score sheets, which differ only in the parent
// they point at.
func (i *dataImport) scores(sheet, parentColumn string, rows []model.ScoreRow, parents map[string]string) []model.ScoreRow {
	seen := map[string]bool{}
	result := make([]model.ScoreRow, 0, len(rows))
	for index, row := range rows {
		entry := index + 1
		parentKey, studentKey := strings.TrimSpace(row.ParentID), strings.TrimSpace(row.StudentID)
		parentID, known := parents[parentKey]
		if !known {
			i.fail(sheet, entry, "%s %q is not listed above", parentColumn, parentKey)
			continue
		}
		studentID, known := i.students[studentKey]
		if !known {
			i.fail(sheet, entry, "studentId %q is not listed in the students sheet", studentKey)
			continue
		}
		if row.Score != nil && *row.Score < 0 {
			i.fail(sheet, entry, "score cannot be negative")
			continue
		}
		pair := parentKey + "\x00" + studentKey
		if seen[pair] {
			continue
		}
		seen[pair] = true
		result = append(result, model.ScoreRow{
			ParentID: parentID, StudentID: studentID, Score: row.Score, RecordedAt: row.RecordedAt,
		})
	}
	return result
}

// adopt makes sure a classroom referenced by a student or a class is offered by
// its school. Real workspaces hold classrooms a school has since stopped
// listing (a student's classroom from an earlier year, say), and a partially
// filled sheet may name a new one, so the school's list grows to match instead
// of the row being rejected.
func (i *dataImport) adopt(schoolKey, classroom string, out *model.Dataset) {
	if containsClassroom(i.classrooms[schoolKey], classroom) {
		return
	}
	i.classrooms[schoolKey] = append(i.classrooms[schoolKey], classroom)
	if index, ok := i.schoolRows[schoolKey]; ok {
		out.Schools[index].Classrooms = append(out.Schools[index].Classrooms, classroom)
	}
}

// link resolves the school and academic year a row belongs to and checks that
// the two match, the same rule the create endpoints enforce.
func (i *dataImport) link(sheet string, entry int, schoolKey, yearKey string) (string, string, bool) {
	schoolID, known := i.schools[schoolKey]
	if !known {
		i.fail(sheet, entry, "schoolId %q is not listed in the schools sheet", schoolKey)
		return "", "", false
	}
	yearID, known := i.years[yearKey]
	if !known {
		i.fail(sheet, entry, "academicYearId %q is not listed in the academic_years sheet", yearKey)
		return "", "", false
	}
	if i.yearSchool[yearKey] != schoolKey {
		i.fail(sheet, entry, "the academic year belongs to a different school")
		return "", "", false
	}
	return schoolID, yearID, true
}
