package service

import (
	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
)

func referencesDTO(items []model.Reference) []dto.Reference {
	result := make([]dto.Reference, len(items))
	for index := range items {
		result[index] = referenceDTO(items[index])
	}
	return result
}

func resultsDTO(items []model.Result) []dto.StudentResult {
	result := make([]dto.StudentResult, len(items))
	for index, item := range items {
		result[index] = dto.StudentResult{
			ID: item.ID, Name: item.Name, Kind: item.Kind, DueAt: item.DueAt,
			Score: item.Score, TotalScore: item.TotalScore,
		}
		if item.CompletedAt != nil {
			result[index].CompletedAt = item.CompletedAt.UTC().Format(timeFormat)
		}
	}
	return result
}

func studentPerformance(item model.Student) dto.Performance {
	all := make([]model.Result, 0, len(item.Assignments)+len(item.Exams))
	all = append(all, item.Assignments...)
	all = append(all, item.Exams...)
	completed := 0
	totalPercent := 0.0
	trend := make([]float64, 0, len(all))
	for _, result := range all {
		if result.Score == nil || result.TotalScore <= 0 {
			continue
		}
		value := *result.Score / result.TotalScore * 100
		totalPercent += value
		completed++
		trend = append(trend, round(value))
	}
	average := 0.0
	if completed > 0 {
		average = round(totalPercent / float64(completed))
	}
	return dto.Performance{
		AverageScore: average, CompletionRate: percent(completed, len(all)),
		Completed: completed, Total: len(all), Standing: standing(average, completed), Trend: trend,
	}
}

func classPerformance(students []model.Student) dto.Performance {
	total := 0
	completed := 0
	totalPercent := 0.0
	trend := make([]float64, 0, len(students))
	for _, student := range students {
		performance := studentPerformance(student)
		total += performance.Total
		completed += performance.Completed
		if performance.Completed > 0 {
			totalPercent += performance.AverageScore
			trend = append(trend, performance.AverageScore)
		}
	}
	average := 0.0
	if len(trend) > 0 {
		average = round(totalPercent / float64(len(trend)))
	}
	return dto.Performance{
		AverageScore: average, CompletionRate: percent(completed, total),
		Completed: completed, Total: total, Standing: standing(average, completed), Trend: trend,
	}
}

func assignmentDTO(item model.Assignment, detail bool) dto.Assignment {
	completed := 0
	for _, student := range item.Students {
		if student.Score != nil {
			completed++
		}
	}
	result := dto.Assignment{
		ID: item.ID, SchoolID: item.SchoolID, AcademicYearID: item.AcademicYearID,
		Name: item.Name, Type: item.Type, DueDate: item.DueDate, TotalScore: item.TotalScore,
		Class: assignmentClassReference(item), AssigneeCount: len(item.Students),
		CompletionCount: completed, Completion: percent(completed, len(item.Students)),
	}
	if !detail {
		return result
	}
	result.Assignees = make([]dto.AssignmentAssignee, len(item.Students))
	for index, assignee := range item.Students {
		result.Assignees[index] = dto.AssignmentAssignee{
			ID: assignee.Student.ID, Name: assignee.Student.FullName(), Grade: assignee.Student.Grade,
			Score: assignee.Score,
		}
		if assignee.CompletedAt != nil {
			result.Assignees[index].CompletedAt = assignee.CompletedAt.UTC().Format(timeFormat)
		}
	}
	return result
}

func examDTO(item model.Exam, detail bool) dto.Exam {
	marked := 0
	scoreTotal := 0.0
	performanceTotal := 0.0
	trend := make([]float64, 0, len(item.Students))
	for _, student := range item.Students {
		if student.Score == nil {
			continue
		}
		marked++
		scoreTotal += *student.Score
		if item.TotalScore > 0 {
			value := *student.Score / item.TotalScore * 100
			performanceTotal += value
			trend = append(trend, round(value))
		}
	}
	averageScore := 0.0
	averagePercent := 0.0
	if marked > 0 {
		averageScore = round(scoreTotal / float64(marked))
		averagePercent = round(performanceTotal / float64(marked))
	}
	result := dto.Exam{
		ID: item.ID, SchoolID: item.SchoolID, AcademicYearID: item.AcademicYearID,
		Name: item.Name, Type: item.Type, ExamDate: item.ExamDate, TotalScore: item.TotalScore,
		Class: referenceDTO(item.Class), StudentCount: len(item.Students), MarkedCount: marked,
		AverageScore: averageScore,
	}
	if !detail {
		return result
	}
	result.Performance = &dto.Performance{
		AverageScore: averagePercent, CompletionRate: percent(marked, len(item.Students)),
		Completed: marked, Total: len(item.Students), Standing: standing(averagePercent, marked), Trend: trend,
	}
	result.Students = make([]dto.ExamStudent, len(item.Students))
	for index, student := range item.Students {
		result.Students[index] = dto.ExamStudent{
			ID: student.Student.ID, Name: student.Student.FullName(), Grade: student.Student.Grade,
			Score: student.Score,
		}
		if student.MarkedAt != nil {
			result.Students[index].MarkedAt = student.MarkedAt.UTC().Format(timeFormat)
		}
	}
	return result
}

func standing(average float64, completed int) string {
	if completed == 0 {
		return "Not marked"
	}
	switch {
	case average >= 85:
		return "Excellent"
	case average >= 70:
		return "On track"
	case average >= 50:
		return "Developing"
	default:
		return "Needs attention"
	}
}

func round(value float64) float64 {
	return float64(int(value*10+0.5)) / 10
}

func assignmentClassReference(item model.Assignment) *dto.Reference {
	if item.Class == nil {
		return nil
	}
	value := referenceDTO(*item.Class)
	return &value
}
