package service

import (
	"context"
	"math"
	"strings"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
)

func (a *App) ListAssignments(ctx context.Context, input dto.ListOptions) (dto.Page[dto.Assignment], error) {
	options, err := listOptions(input, map[string]bool{
		"name": true, "type": true, "dueDate": true, "completion": true,
	}, "dueDate")
	if err != nil {
		return dto.Page[dto.Assignment]{}, err
	}
	items, total, err := a.storage.ListAssignments(ctx, options)
	if err != nil {
		return dto.Page[dto.Assignment]{}, err
	}
	result := make([]dto.Assignment, len(items))
	for index := range items {
		result[index] = assignmentDTO(items[index], false, "")
	}
	return dto.Page[dto.Assignment]{Items: result, Total: int(total), Page: options.Page, PageSize: options.PageSize}, nil
}

func (a *App) GetAssignment(ctx context.Context, id string) (dto.Assignment, error) {
	return a.GetAssignmentFiltered(ctx, id, "")
}

func (a *App) GetAssignmentFiltered(ctx context.Context, id, classroom string) (dto.Assignment, error) {
	item, err := a.storage.Assignment(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Assignment{}, translate(err, "assignment")
	}
	return assignmentDTO(item, true, strings.TrimSpace(classroom)), nil
}

func (a *App) CreateAssignment(ctx context.Context, input dto.AssignmentRequest) (dto.Assignment, error) {
	item, studentIDs, err := a.assignmentFromInput(ctx, "", input)
	if err != nil {
		return dto.Assignment{}, err
	}
	item.ID, err = a.newID("asg")
	if err != nil {
		return dto.Assignment{}, err
	}
	created, err := a.storage.CreateAssignment(ctx, item, studentIDs)
	if err != nil {
		return dto.Assignment{}, translate(err, "assignment")
	}
	return assignmentDTO(created, true, ""), nil
}

func (a *App) UpdateAssignment(ctx context.Context, id string, input dto.UpdateAssignmentRequest) (dto.Assignment, error) {
	current, err := a.storage.Assignment(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Assignment{}, translate(err, "assignment")
	}
	merged := assignmentRequest(current)
	if input.SchoolID != nil {
		merged.SchoolID = *input.SchoolID
	}
	if input.AcademicYearID != nil {
		merged.AcademicYearID = *input.AcademicYearID
	}
	if input.Name != nil {
		merged.Name = *input.Name
	}
	if input.Type != nil {
		merged.Type = *input.Type
	}
	if input.DueDate != nil {
		merged.DueDate = *input.DueDate
	}
	if input.TotalScore != nil {
		merged.TotalScore = *input.TotalScore
	}
	if input.ClassID != nil {
		merged.ClassID = *input.ClassID
	}
	if input.StudentIDs != nil {
		merged.StudentIDs = *input.StudentIDs
	}
	item, studentIDs, err := a.assignmentFromInput(ctx, current.ID, merged)
	if err != nil {
		return dto.Assignment{}, err
	}
	rosterSupplied := input.StudentIDs != nil || input.ClassID != nil || input.Type != nil
	var roster *[]string
	if rosterSupplied {
		roster = &studentIDs
	}
	updated, err := a.storage.UpdateAssignment(ctx, item, roster)
	if err != nil {
		return dto.Assignment{}, translate(err, "assignment")
	}
	return assignmentDTO(updated, true, ""), nil
}

func (a *App) DeleteAssignment(ctx context.Context, id string) error {
	assignment, err := a.storage.Assignment(ctx, strings.TrimSpace(id))
	if err != nil {
		return translate(err, "assignment")
	}
	for _, assignee := range assignment.Students {
		if assignee.Score != nil || assignee.CompletedAt != nil {
			return conflict("assignment cannot be deleted after a score is recorded")
		}
	}
	return translate(a.storage.DeleteAssignment(ctx, assignment.ID), "assignment")
}

func (a *App) SetAssignmentScore(ctx context.Context, assignmentID, studentID string, input dto.ScoreRequest) (dto.Assignment, error) {
	assignment, err := a.storage.Assignment(ctx, strings.TrimSpace(assignmentID))
	if err != nil {
		return dto.Assignment{}, translate(err, "assignment")
	}
	if !validScore(input.Score, assignment.TotalScore) {
		return dto.Assignment{}, validation("score must be between zero and totalScore")
	}
	now := a.now().UTC()
	if err := a.storage.SetAssignmentScore(ctx, assignment.ID, strings.TrimSpace(studentID), &input.Score, &now); err != nil {
		return dto.Assignment{}, translate(err, "assignment assignee")
	}
	updated, err := a.storage.Assignment(ctx, assignment.ID)
	if err != nil {
		return dto.Assignment{}, translate(err, "assignment")
	}
	return assignmentDTO(updated, true, ""), nil
}

func (a *App) ClearAssignmentScore(ctx context.Context, assignmentID, studentID string) error {
	err := a.storage.SetAssignmentScore(ctx, strings.TrimSpace(assignmentID), strings.TrimSpace(studentID), nil, nil)
	return translate(err, "assignment assignee")
}

func (a *App) assignmentFromInput(ctx context.Context, id string, input dto.AssignmentRequest) (model.Assignment, []string, error) {
	input.SchoolID = strings.TrimSpace(input.SchoolID)
	input.AcademicYearID = strings.TrimSpace(input.AcademicYearID)
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	input.ClassID = strings.TrimSpace(input.ClassID)
	switch {
	case input.SchoolID == "":
		return model.Assignment{}, nil, validation("schoolId is required")
	case input.AcademicYearID == "":
		return model.Assignment{}, nil, validation("academicYearId is required")
	case input.Name == "":
		return model.Assignment{}, nil, validation("name is required")
	case input.Type != "class" && input.Type != "individual":
		return model.Assignment{}, nil, validation("type must be class or individual")
	case !finitePositive(input.TotalScore):
		return model.Assignment{}, nil, validation("totalScore must be positive")
	}
	dueDate, err := date(input.DueDate, "dueDate")
	if err != nil {
		return model.Assignment{}, nil, err
	}
	year, err := a.storage.AcademicYear(ctx, input.AcademicYearID)
	if err != nil {
		return model.Assignment{}, nil, translate(err, "academic year")
	}
	if year.SchoolID != input.SchoolID {
		return model.Assignment{}, nil, &Error{Kind: ErrConflict, Message: "school and academic year do not match"}
	}
	studentIDs, err := uniqueTrimmed(input.StudentIDs, "studentIds")
	if err != nil {
		return model.Assignment{}, nil, err
	}
	var classID *string
	if input.Type == "class" {
		if input.ClassID == "" {
			return model.Assignment{}, nil, validation("classId is required for class assignments")
		}
		if len(studentIDs) > 0 {
			return model.Assignment{}, nil, validation("studentIds are not allowed for class assignments")
		}
		classItem, findErr := a.storage.Class(ctx, input.ClassID)
		if findErr != nil {
			return model.Assignment{}, nil, translate(findErr, "class")
		}
		if classItem.SchoolID != input.SchoolID || classItem.AcademicYearID != input.AcademicYearID {
			return model.Assignment{}, nil, &Error{Kind: ErrConflict, Message: "class does not belong to the school and academic year"}
		}
		classID = &input.ClassID
	} else {
		if input.ClassID != "" {
			return model.Assignment{}, nil, validation("classId is not allowed for individual assignments")
		}
		if len(studentIDs) == 0 {
			return model.Assignment{}, nil, validation("at least one studentId is required for individual assignments")
		}
		for _, studentID := range studentIDs {
			student, findErr := a.storage.Student(ctx, studentID)
			if findErr != nil {
				return model.Assignment{}, nil, translate(findErr, "student")
			}
			if student.SchoolID != input.SchoolID {
				return model.Assignment{}, nil, &Error{Kind: ErrConflict, Message: "student belongs to a different school"}
			}
		}
	}
	return model.Assignment{
		ID: id, SchoolID: input.SchoolID, AcademicYearID: input.AcademicYearID,
		Name: input.Name, Type: input.Type, DueDate: dueDate.Format("2006-01-02"),
		TotalScore: input.TotalScore, ClassID: classID,
	}, studentIDs, nil
}

func assignmentRequest(item model.Assignment) dto.AssignmentRequest {
	classID := ""
	if item.ClassID != nil {
		classID = *item.ClassID
	}
	studentIDs := make([]string, 0, len(item.Students))
	if item.Type == "individual" {
		for _, student := range item.Students {
			studentIDs = append(studentIDs, student.Student.ID)
		}
	}
	return dto.AssignmentRequest{
		SchoolID: item.SchoolID, AcademicYearID: item.AcademicYearID, Name: item.Name,
		Type: item.Type, DueDate: item.DueDate, TotalScore: item.TotalScore,
		ClassID: classID, StudentIDs: studentIDs,
	}
}

func (a *App) ListExams(ctx context.Context, input dto.ListOptions) (dto.Page[dto.Exam], error) {
	options, err := listOptions(input, map[string]bool{
		"name": true, "type": true, "examDate": true, "markedCount": true, "averageScore": true,
	}, "examDate")
	if err != nil {
		return dto.Page[dto.Exam]{}, err
	}
	items, total, err := a.storage.ListExams(ctx, options)
	if err != nil {
		return dto.Page[dto.Exam]{}, err
	}
	result := make([]dto.Exam, len(items))
	for index := range items {
		result[index] = examDTO(items[index], false, "")
	}
	return dto.Page[dto.Exam]{Items: result, Total: int(total), Page: options.Page, PageSize: options.PageSize}, nil
}

func (a *App) GetExam(ctx context.Context, id string) (dto.Exam, error) {
	return a.GetExamFiltered(ctx, id, "")
}

func (a *App) GetExamFiltered(ctx context.Context, id, classroom string) (dto.Exam, error) {
	item, err := a.storage.Exam(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Exam{}, translate(err, "exam")
	}
	return examDTO(item, true, strings.TrimSpace(classroom)), nil
}

func (a *App) CreateExam(ctx context.Context, input dto.ExamRequest) (dto.Exam, error) {
	item, err := a.examFromInput(ctx, "", input)
	if err != nil {
		return dto.Exam{}, err
	}
	item.ID, err = a.newID("exam")
	if err != nil {
		return dto.Exam{}, err
	}
	created, err := a.storage.CreateExam(ctx, item)
	if err != nil {
		return dto.Exam{}, translate(err, "exam")
	}
	return examDTO(created, true, ""), nil
}

func (a *App) UpdateExam(ctx context.Context, id string, input dto.UpdateExamRequest) (dto.Exam, error) {
	current, err := a.storage.Exam(ctx, strings.TrimSpace(id))
	if err != nil {
		return dto.Exam{}, translate(err, "exam")
	}
	merged := examRequest(current)
	if input.SchoolID != nil {
		merged.SchoolID = *input.SchoolID
	}
	if input.AcademicYearID != nil {
		merged.AcademicYearID = *input.AcademicYearID
	}
	if input.ClassID != nil {
		merged.ClassID = *input.ClassID
	}
	if input.Name != nil {
		merged.Name = *input.Name
	}
	if input.Type != nil {
		merged.Type = *input.Type
	}
	if input.ExamDate != nil {
		merged.ExamDate = *input.ExamDate
	}
	if input.TotalScore != nil {
		merged.TotalScore = *input.TotalScore
	}
	item, err := a.examFromInput(ctx, current.ID, merged)
	if err != nil {
		return dto.Exam{}, err
	}
	updated, err := a.storage.UpdateExam(ctx, item)
	if err != nil {
		return dto.Exam{}, translate(err, "exam")
	}
	return examDTO(updated, true, ""), nil
}

func (a *App) DeleteExam(ctx context.Context, id string) error {
	exam, err := a.storage.Exam(ctx, strings.TrimSpace(id))
	if err != nil {
		return translate(err, "exam")
	}
	for _, student := range exam.Students {
		if student.Score != nil || student.MarkedAt != nil {
			return conflict("exam cannot be deleted after a score is recorded")
		}
	}
	return translate(a.storage.DeleteExam(ctx, exam.ID), "exam")
}

func (a *App) SetExamScore(ctx context.Context, examID, studentID string, input dto.ScoreRequest) (dto.Exam, error) {
	exam, err := a.storage.Exam(ctx, strings.TrimSpace(examID))
	if err != nil {
		return dto.Exam{}, translate(err, "exam")
	}
	if !validScore(input.Score, exam.TotalScore) {
		return dto.Exam{}, validation("score must be between zero and totalScore")
	}
	now := a.now().UTC()
	if err := a.storage.SetExamScore(ctx, exam.ID, strings.TrimSpace(studentID), &input.Score, &now); err != nil {
		return dto.Exam{}, translate(err, "exam student")
	}
	updated, err := a.storage.Exam(ctx, exam.ID)
	if err != nil {
		return dto.Exam{}, translate(err, "exam")
	}
	return examDTO(updated, true, ""), nil
}

func (a *App) ClearExamScore(ctx context.Context, examID, studentID string) error {
	err := a.storage.SetExamScore(ctx, strings.TrimSpace(examID), strings.TrimSpace(studentID), nil, nil)
	return translate(err, "exam student")
}

func (a *App) examFromInput(ctx context.Context, id string, input dto.ExamRequest) (model.Exam, error) {
	input.SchoolID = strings.TrimSpace(input.SchoolID)
	input.AcademicYearID = strings.TrimSpace(input.AcademicYearID)
	input.ClassID = strings.TrimSpace(input.ClassID)
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.TrimSpace(input.Type)
	switch {
	case input.SchoolID == "":
		return model.Exam{}, validation("schoolId is required")
	case input.AcademicYearID == "":
		return model.Exam{}, validation("academicYearId is required")
	case input.ClassID == "":
		return model.Exam{}, validation("classId is required")
	case input.Name == "":
		return model.Exam{}, validation("name is required")
	case input.Type == "":
		return model.Exam{}, validation("type is required")
	case !finitePositive(input.TotalScore):
		return model.Exam{}, validation("totalScore must be positive")
	}
	examDate, err := date(input.ExamDate, "examDate")
	if err != nil {
		return model.Exam{}, err
	}
	year, err := a.storage.AcademicYear(ctx, input.AcademicYearID)
	if err != nil {
		return model.Exam{}, translate(err, "academic year")
	}
	classItem, err := a.storage.Class(ctx, input.ClassID)
	if err != nil {
		return model.Exam{}, translate(err, "class")
	}
	if year.SchoolID != input.SchoolID || classItem.SchoolID != input.SchoolID ||
		classItem.AcademicYearID != input.AcademicYearID {
		return model.Exam{}, &Error{Kind: ErrConflict, Message: "school, academic year, and class do not match"}
	}
	return model.Exam{
		ID: id, SchoolID: input.SchoolID, AcademicYearID: input.AcademicYearID, ClassID: input.ClassID,
		Name: input.Name, Type: input.Type, ExamDate: examDate.Format("2006-01-02"), TotalScore: input.TotalScore,
	}, nil
}

func examRequest(item model.Exam) dto.ExamRequest {
	return dto.ExamRequest{
		SchoolID: item.SchoolID, AcademicYearID: item.AcademicYearID, ClassID: item.ClassID,
		Name: item.Name, Type: item.Type, ExamDate: item.ExamDate, TotalScore: item.TotalScore,
	}
}

func finitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validScore(score, total float64) bool {
	return score >= 0 && score <= total && !math.IsNaN(score) && !math.IsInf(score, 0)
}
