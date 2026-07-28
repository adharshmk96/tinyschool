package service

import (
	"context"
	"fmt"
	"strings"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
)

func (a *App) ListClasses(ctx context.Context, input dto.ListOptions) (dto.Page[dto.Class], error) {
	options, err := listOptions(input, map[string]bool{
		"name": true, "subject": true, "grade": true, "studentCount": true,
	}, "name")
	if err != nil {
		return dto.Page[dto.Class]{}, err
	}
	items, total, err := a.storage.ListClasses(ctx, options)
	if err != nil {
		return dto.Page[dto.Class]{}, err
	}
	result := make([]dto.Class, len(items))
	for index := range items {
		result[index] = classDTO(items[index], false)
	}
	return dto.Page[dto.Class]{Items: result, Total: int(total), Page: options.Page, PageSize: options.PageSize}, nil
}

func (a *App) GetClass(ctx context.Context, id string) (dto.Class, error) {
	item, err := a.storage.Class(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Class{}, translate(err, "class")
	}
	return classDTO(item, true), nil
}

func (a *App) CreateClass(ctx context.Context, input dto.ClassRequest) (dto.Class, error) {
	item, studentIDs, err := a.classFromInput(ctx, "", input)
	if err != nil {
		return dto.Class{}, err
	}
	item.ID, err = a.newID("cls")
	if err != nil {
		return dto.Class{}, err
	}
	created, err := a.storage.CreateClass(ctx, item, studentIDs)
	if err != nil {
		return dto.Class{}, translate(err, "class")
	}
	return classDTO(created, true), nil
}

func (a *App) UpdateClass(ctx context.Context, id string, input dto.UpdateClassRequest) (dto.Class, error) {
	current, err := a.storage.Class(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Class{}, translate(err, "class")
	}
	merged := dto.ClassRequest{
		SchoolID: current.SchoolID, AcademicYearID: current.AcademicYearID, Name: current.Name,
		Subject: current.Subject, Grade: current.Grade, Description: current.Description,
	}
	if input.SchoolID != nil {
		merged.SchoolID = *input.SchoolID
	}
	if input.AcademicYearID != nil {
		merged.AcademicYearID = *input.AcademicYearID
	}
	if input.Name != nil {
		merged.Name = *input.Name
	}
	if input.Subject != nil {
		merged.Subject = *input.Subject
	}
	if input.Grade != nil {
		merged.Grade = *input.Grade
	}
	if input.Description != nil {
		merged.Description = *input.Description
	}
	if input.StudentIDs != nil {
		merged.StudentIDs = *input.StudentIDs
	}
	item, studentIDs, err := a.classFromInput(ctx, current.ID, merged)
	if err != nil {
		return dto.Class{}, err
	}
	updated, err := a.storage.UpdateClass(ctx, item, optionalIDs(input.StudentIDs, studentIDs))
	if err != nil {
		return dto.Class{}, translate(err, "class")
	}
	return classDTO(updated, true), nil
}

func (a *App) DeleteClass(ctx context.Context, id string) error {
	return translate(a.storage.DeleteClass(ctx, strings.TrimSpace(id)), "class")
}

func (a *App) classFromInput(ctx context.Context, id string, input dto.ClassRequest) (model.Class, []string, error) {
	input.SchoolID = strings.TrimSpace(input.SchoolID)
	input.AcademicYearID = strings.TrimSpace(input.AcademicYearID)
	input.Name = strings.TrimSpace(input.Name)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Grade = strings.TrimSpace(input.Grade)
	input.Description = strings.TrimSpace(input.Description)
	switch {
	case input.SchoolID == "":
		return model.Class{}, nil, validation("schoolId is required")
	case input.AcademicYearID == "":
		return model.Class{}, nil, validation("academicYearId is required")
	case input.Name == "":
		return model.Class{}, nil, validation("name is required")
	case input.Subject == "":
		return model.Class{}, nil, validation("subject is required")
	case input.Grade == "":
		return model.Class{}, nil, validation("grade is required")
	}
	year, err := a.storage.AcademicYear(ctx, input.AcademicYearID)
	if err != nil {
		return model.Class{}, nil, translate(err, "academic year")
	}
	if year.SchoolID != input.SchoolID {
		return model.Class{}, nil, &Error{Kind: ErrConflict, Message: "school and academic year do not match"}
	}
	studentIDs, err := uniqueTrimmed(input.StudentIDs, "studentIds")
	if err != nil {
		return model.Class{}, nil, err
	}
	for _, studentID := range studentIDs {
		student, findErr := a.storage.Student(ctx, studentID)
		if findErr != nil {
			return model.Class{}, nil, translate(findErr, "student")
		}
		if student.SchoolID != input.SchoolID {
			return model.Class{}, nil, &Error{Kind: ErrConflict, Message: "student belongs to a different school"}
		}
	}
	return model.Class{
		ID: id, SchoolID: input.SchoolID, AcademicYearID: input.AcademicYearID,
		Name: input.Name, Subject: input.Subject, Grade: input.Grade, Description: input.Description,
	}, studentIDs, nil
}

func optionalIDs(input *[]string, normalized []string) *[]string {
	if input == nil {
		return nil
	}
	return &normalized
}

func (a *App) ListStudents(ctx context.Context, input dto.ListOptions) (dto.Page[dto.Student], error) {
	options, err := listOptions(input, map[string]bool{
		"name": true, "email": true, "grade": true, "averageScore": true,
	}, "name")
	if err != nil {
		return dto.Page[dto.Student]{}, err
	}
	items, total, err := a.storage.ListStudents(ctx, options)
	if err != nil {
		return dto.Page[dto.Student]{}, err
	}
	result := make([]dto.Student, len(items))
	for index := range items {
		result[index] = studentDTO(items[index], false)
	}
	return dto.Page[dto.Student]{Items: result, Total: int(total), Page: options.Page, PageSize: options.PageSize}, nil
}

func (a *App) GetStudent(ctx context.Context, id string) (dto.Student, error) {
	item, err := a.storage.Student(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Student{}, translate(err, "student")
	}
	return studentDTO(item, true), nil
}

func (a *App) CreateStudent(ctx context.Context, input dto.StudentRequest) (dto.Student, error) {
	item, err := a.studentFromInput("", input)
	if err != nil {
		return dto.Student{}, err
	}
	if _, err := a.storage.School(ctx, item.SchoolID); err != nil {
		return dto.Student{}, translate(err, "school")
	}
	item.ID, err = a.newID("stu")
	if err != nil {
		return dto.Student{}, err
	}
	created, err := a.storage.CreateStudent(ctx, item)
	if err != nil {
		return dto.Student{}, translate(err, "student")
	}
	return studentDTO(created, true), nil
}

func (a *App) UpdateStudent(ctx context.Context, id string, input dto.UpdateStudentRequest) (dto.Student, error) {
	current, err := a.storage.Student(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Student{}, translate(err, "student")
	}
	merged := studentRequest(current)
	if input.SchoolID != nil {
		merged.SchoolID = *input.SchoolID
	}
	if input.FirstName != nil {
		merged.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		merged.LastName = *input.LastName
	}
	if input.Email != nil {
		merged.Email = *input.Email
	}
	if input.Phone != nil {
		merged.Phone = *input.Phone
	}
	if input.Grade != nil {
		merged.Grade = *input.Grade
	}
	if input.Guardian != nil {
		merged.Guardian = *input.Guardian
	}
	if input.ResidentAddress != nil {
		merged.ResidentAddress = *input.ResidentAddress
	}
	if input.PermanentAddress != nil {
		merged.PermanentAddress = *input.PermanentAddress
	}
	item, err := a.studentFromInput(current.ID, merged)
	if err != nil {
		return dto.Student{}, err
	}
	if _, err := a.storage.School(ctx, item.SchoolID); err != nil {
		return dto.Student{}, translate(err, "school")
	}
	updated, err := a.storage.UpdateStudent(ctx, item)
	if err != nil {
		return dto.Student{}, translate(err, "student")
	}
	return studentDTO(updated, true), nil
}

func (a *App) DeleteStudent(ctx context.Context, id string) error {
	student, err := a.storage.Student(ctx, strings.TrimSpace(id))
	if err != nil {
		return translate(err, "student")
	}

	var conflicts []string
	if len(student.Classes) > 0 {
		conflicts = append(conflicts, namedConflicts("class", student.Classes))
	}
	if len(student.Assignments) > 0 {
		conflicts = append(conflicts, resultConflicts("assignment", student.Assignments))
	}
	if len(student.Exams) > 0 {
		conflicts = append(conflicts, resultConflicts("exam", student.Exams))
	}
	notes, behaviourLogs := 0, 0
	for _, log := range student.Logs {
		if log.Kind == "note" {
			notes++
		} else {
			behaviourLogs++
		}
	}
	if notes > 0 {
		conflicts = append(conflicts, countConflict("note", notes))
	}
	if behaviourLogs > 0 {
		conflicts = append(conflicts, countConflict("behaviour log", behaviourLogs))
	}
	if len(conflicts) > 0 {
		return conflict("student cannot be deleted because it has related data: " + strings.Join(conflicts, "; "))
	}
	return translate(a.storage.DeleteStudent(ctx, student.ID), "student")
}

func namedConflicts(kind string, items []model.Reference) string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, `"`+item.Name+`"`)
	}
	if len(items) > 1 {
		if kind == "class" {
			kind = "classes"
		} else {
			kind += "s"
		}
	}
	return kind + " " + strings.Join(names, ", ")
}

func resultConflicts(kind string, items []model.Result) string {
	references := make([]model.Reference, 0, len(items))
	for _, item := range items {
		references = append(references, model.Reference{ID: item.ID, Name: item.Name})
	}
	return namedConflicts(kind, references)
}

func countConflict(kind string, count int) string {
	if count > 1 {
		kind += "s"
	}
	return fmt.Sprintf("%d %s", count, kind)
}

func (a *App) AddBehaviour(ctx context.Context, studentID string, input dto.BehaviourRequest) (dto.StudentLog, error) {
	kind := strings.ToLower(strings.TrimSpace(input.Type))
	if kind != "positive" && kind != "need_attention" && kind != "incident" {
		return dto.StudentLog{}, validation("type must be positive, need_attention, or incident")
	}
	return a.createLog(ctx, studentID, "behaviour", kind, input.Note, "beh")
}

func (a *App) DeleteBehaviour(ctx context.Context, studentID, logID string) error {
	return translate(a.storage.DeleteStudentLog(ctx, strings.TrimSpace(studentID), strings.TrimSpace(logID), "behaviour"), "behaviour log")
}

func (a *App) AddNote(ctx context.Context, studentID string, input dto.NoteRequest) (dto.StudentLog, error) {
	return a.createLog(ctx, studentID, "note", "", input.Note, "note")
}

func (a *App) DeleteNote(ctx context.Context, studentID, logID string) error {
	return translate(a.storage.DeleteStudentLog(ctx, strings.TrimSpace(studentID), strings.TrimSpace(logID), "note"), "note")
}

func (a *App) createLog(ctx context.Context, studentID, kind, logType, note, prefix string) (dto.StudentLog, error) {
	studentID = strings.TrimSpace(studentID)
	note = strings.TrimSpace(note)
	if note == "" {
		return dto.StudentLog{}, validation("note is required")
	}
	if _, err := a.storage.Student(ctx, studentID); err != nil {
		return dto.StudentLog{}, translate(err, "student")
	}
	id, err := a.newID(prefix)
	if err != nil {
		return dto.StudentLog{}, err
	}
	created, err := a.storage.CreateStudentLog(ctx, model.StudentLog{
		ID: id, StudentID: studentID, Kind: kind, Type: logType, Note: note, CreatedAt: a.now().UTC(),
	})
	if err != nil {
		return dto.StudentLog{}, translate(err, kind)
	}
	return studentLogDTO(created), nil
}

func (a *App) studentFromInput(id string, input dto.StudentRequest) (model.Student, error) {
	input.SchoolID = strings.TrimSpace(input.SchoolID)
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	input.Grade = strings.TrimSpace(input.Grade)
	input.Guardian.Name = strings.TrimSpace(input.Guardian.Name)
	input.Guardian.Email = strings.ToLower(strings.TrimSpace(input.Guardian.Email))
	input.Guardian.Phone = strings.TrimSpace(input.Guardian.Phone)
	input.ResidentAddress = strings.TrimSpace(input.ResidentAddress)
	input.PermanentAddress = strings.TrimSpace(input.PermanentAddress)
	switch {
	case input.SchoolID == "":
		return model.Student{}, validation("schoolId is required")
	case input.FirstName == "":
		return model.Student{}, validation("firstName is required")
	case input.LastName == "":
		return model.Student{}, validation("lastName is required")
	case input.Grade == "":
		return model.Student{}, validation("grade is required")
	}
	if input.Email != "" {
		if err := validEmail(input.Email); err != nil {
			return model.Student{}, err
		}
	}
	if input.Guardian.Email != "" {
		if err := validEmail(input.Guardian.Email); err != nil {
			return model.Student{}, validation("guardian email must be valid")
		}
	}
	return model.Student{
		ID: id, SchoolID: input.SchoolID, FirstName: input.FirstName, LastName: input.LastName,
		Email: input.Email, Phone: input.Phone, Grade: input.Grade,
		GuardianName: input.Guardian.Name, GuardianEmail: input.Guardian.Email, GuardianPhone: input.Guardian.Phone,
		ResidentAddress: input.ResidentAddress, PermanentAddress: input.PermanentAddress,
	}, nil
}

func studentRequest(item model.Student) dto.StudentRequest {
	return dto.StudentRequest{
		SchoolID: item.SchoolID, FirstName: item.FirstName, LastName: item.LastName,
		Email: item.Email, Phone: item.Phone, Grade: item.Grade,
		Guardian:        dto.Guardian{Name: item.GuardianName, Email: item.GuardianEmail, Phone: item.GuardianPhone},
		ResidentAddress: item.ResidentAddress, PermanentAddress: item.PermanentAddress,
	}
}

func classDTO(item model.Class, detail bool) dto.Class {
	performance := classPerformance(item.Students)
	result := dto.Class{
		ID: item.ID, SchoolID: item.SchoolID, AcademicYearID: item.AcademicYearID,
		Name: item.Name, Subject: item.Subject, Grade: item.Grade, Description: item.Description,
		StudentCount: len(item.Students), AverageScore: performance.AverageScore,
	}
	if !detail {
		return result
	}
	result.Performance = &performance
	result.Students = make([]dto.Reference, len(item.Students))
	for index, student := range item.Students {
		result.Students[index] = dto.Reference{ID: student.ID, Name: student.FullName()}
	}
	result.Assignments = referencesDTO(item.Assignments)
	result.Exams = referencesDTO(item.Exams)
	return result
}

func studentDTO(item model.Student, detail bool) dto.Student {
	performance := studentPerformance(item)
	result := dto.Student{
		ID: item.ID, SchoolID: item.SchoolID, FirstName: item.FirstName, LastName: item.LastName,
		FullName: item.FullName(), Email: item.Email, Phone: item.Phone, Grade: item.Grade,
		Guardian:        dto.Guardian{Name: item.GuardianName, Email: item.GuardianEmail, Phone: item.GuardianPhone},
		ResidentAddress: item.ResidentAddress, PermanentAddress: item.PermanentAddress,
		AverageScore: performance.AverageScore, ClassAverage: performance.ClassAverage,
	}
	if !detail {
		return result
	}
	result.Performance = &performance
	result.Behaviour = []dto.StudentLog{}
	result.Notes = []dto.StudentLog{}
	for _, log := range item.Logs {
		if log.Kind == "behaviour" {
			result.Behaviour = append(result.Behaviour, studentLogDTO(log))
		} else if log.Kind == "note" {
			result.Notes = append(result.Notes, studentLogDTO(log))
		}
	}
	result.Assignments = resultsDTO(item.Assignments)
	result.Exams = resultsDTO(item.Exams)
	return result
}

func studentLogDTO(item model.StudentLog) dto.StudentLog {
	return dto.StudentLog{
		ID: item.ID, Type: item.Type, Note: item.Note, CreatedAt: item.CreatedAt.UTC().Format(timeFormat),
	}
}
