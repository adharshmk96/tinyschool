package service

import (
	"math"
	"sort"
	"strings"
	"time"

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
	performance := dto.Performance{
		AverageScore: average, CompletionRate: percent(completed, len(all)),
		Completed: completed, Total: len(all), Standing: standing(average, completed), Trend: trend,
	}
	assignmentMonths := map[string][]float64{}
	examMonths := map[string][]float64{}
	assignmentTotal := 0.0
	examTotal := 0.0
	allScores := make([]float64, 0, completed)
	for _, result := range item.Assignments {
		performance.AssignmentTotal++
		if result.Score == nil || result.TotalScore <= 0 {
			continue
		}
		score := round(*result.Score / result.TotalScore * 100)
		performance.AssignmentCompleted++
		assignmentTotal += score
		allScores = append(allScores, score)
		addMonthlyScore(assignmentMonths, result.DueAt, score)
		if performance.BestResult == nil || score > performance.BestResult.Score {
			performance.BestResult = &dto.PerformanceResult{Name: result.Name, Kind: "Assignment", Score: score}
		}
	}
	for _, result := range item.Exams {
		performance.ExamTotal++
		if result.Score == nil || result.TotalScore <= 0 {
			continue
		}
		score := round(*result.Score / result.TotalScore * 100)
		performance.ExamCompleted++
		examTotal += score
		allScores = append(allScores, score)
		addMonthlyScore(examMonths, result.DueAt, score)
		if performance.BestResult == nil || score > performance.BestResult.Score {
			performance.BestResult = &dto.PerformanceResult{Name: result.Name, Kind: "Exam", Score: score}
		}
	}
	if performance.AssignmentCompleted > 0 {
		performance.AssignmentAverage = round(assignmentTotal / float64(performance.AssignmentCompleted))
	}
	if performance.ExamCompleted > 0 {
		performance.ExamAverage = round(examTotal / float64(performance.ExamCompleted))
	}
	performance.AssignmentCompletion = percent(performance.AssignmentCompleted, performance.AssignmentTotal)
	performance.ExamCompletion = percent(performance.ExamCompleted, performance.ExamTotal)
	performance.MonthlyAssignmentTrend = monthlyTrend(assignmentMonths)
	performance.MonthlyExamTrend = monthlyTrend(examMonths)
	switch {
	case performance.AssignmentCompleted == 0 && performance.ExamCompleted == 0:
		performance.StrongestArea = "Waiting for scores"
	case performance.ExamCompleted == 0 || performance.AssignmentAverage > performance.ExamAverage:
		performance.StrongestArea = "Assignments"
	case performance.AssignmentCompleted == 0 || performance.ExamAverage > performance.AssignmentAverage:
		performance.StrongestArea = "Exams"
	default:
		performance.StrongestArea = "Balanced"
	}
	if len(allScores) > 0 {
		variance := 0.0
		for _, score := range allScores {
			difference := score - average
			variance += difference * difference
		}
		deviation := math.Sqrt(variance / float64(len(allScores)))
		performance.ScoreConsistency = max(0, min(100, int(math.Round(100-deviation))))
	}
	return performance
}

func addMonthlyScore(months map[string][]float64, date string, score float64) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return
	}
	months[date[:7]] = append(months[date[:7]], score)
}

func monthlyTrend(monthScores map[string][]float64) []dto.TrendPoint {
	months := make([]string, 0, len(monthScores))
	for month := range monthScores {
		months = append(months, month)
	}
	sort.Strings(months)
	trend := make([]dto.TrendPoint, 0, len(months))
	for _, month := range months {
		total := 0.0
		for _, value := range monthScores[month] {
			total += value
		}
		parsed, _ := time.Parse("2006-01", month)
		trend = append(trend, dto.TrendPoint{
			Label: parsed.Format("Jan 2006"),
			Value: round(total / float64(len(monthScores[month]))),
		})
	}
	return trend
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
	result := dto.Performance{
		AverageScore: average, CompletionRate: percent(completed, total),
		Completed: completed, Total: total, Standing: standing(average, completed), Trend: trend,
	}
	result.TopStudents = rankedStudents(students)
	monthScores := map[string][]float64{}
	for _, student := range students {
		for _, assignment := range student.Assignments {
			result.AssignmentTotal++
			if assignment.Score != nil {
				result.AssignmentCompleted++
			}
		}
		for _, exam := range student.Exams {
			result.ExamTotal++
			if exam.Score == nil || exam.TotalScore <= 0 {
				continue
			}
			result.ExamCompleted++
			if _, err := time.Parse("2006-01-02", exam.DueAt); err == nil {
				month := exam.DueAt[:7]
				monthScores[month] = append(monthScores[month], *exam.Score/exam.TotalScore*100)
			}
		}
	}
	result.AssignmentCompletion = percent(result.AssignmentCompleted, result.AssignmentTotal)
	result.ExamCompletion = percent(result.ExamCompleted, result.ExamTotal)
	months := make([]string, 0, len(monthScores))
	for month := range monthScores {
		months = append(months, month)
	}
	sort.Strings(months)
	for _, month := range months {
		values := monthScores[month]
		total := 0.0
		for _, value := range values {
			total += value
		}
		parsed, _ := time.Parse("2006-01", month)
		result.MonthlyExamTrend = append(result.MonthlyExamTrend, dto.TrendPoint{
			Label: parsed.Format("Jan 2006"), Value: round(total / float64(len(values))),
		})
	}
	return result
}

func assignmentDTO(item model.Assignment, detail bool, classroomFilter string) dto.Assignment {
	students := item.Students
	if detail && classroomFilter != "" {
		filtered := make([]model.AssignmentStudent, 0, len(students))
		for _, assignee := range students {
			if strings.EqualFold(strings.TrimSpace(assignee.Student.ClassroomFor(item.AcademicYearID)), strings.TrimSpace(classroomFilter)) {
				filtered = append(filtered, assignee)
			}
		}
		students = filtered
	}
	completed := 0
	for _, student := range students {
		if student.Score != nil {
			completed++
		}
	}
	result := dto.Assignment{
		ID: item.ID, SchoolID: item.SchoolID, AcademicYearID: item.AcademicYearID,
		Name: item.Name, Type: item.Type, DueDate: item.DueDate, TotalScore: item.TotalScore,
		Class: assignmentClassReference(item), AssigneeCount: len(students),
		CompletionCount: completed, Completion: percent(completed, len(students)),
	}
	if !detail {
		return result
	}
	scores := make([]float64, 0, len(students))
	ranked := make([]dto.RankedStudent, 0, len(students))
	scoreTotal := 0.0
	for _, assignee := range students {
		if assignee.Score == nil || item.TotalScore <= 0 {
			continue
		}
		value := round(*assignee.Score / item.TotalScore * 100)
		scores = append(scores, value)
		scoreTotal += value
		ranked = append(ranked, dto.RankedStudent{
			ID: assignee.Student.ID, Name: assignee.Student.FullName(), Score: value,
		})
	}
	average := 0.0
	if len(scores) > 0 {
		average = round(scoreTotal / float64(len(scores)))
	}
	result.Performance = &dto.Performance{
		AverageScore: average, CompletionRate: percent(completed, len(students)),
		Completed: completed, Total: len(students), Standing: standing(average, completed),
	}
	sort.Float64s(scores)
	if len(scores) > 0 {
		result.Performance.LowestScore = scores[0]
		result.Performance.HighestScore = scores[len(scores)-1]
		middle := len(scores) / 2
		result.Performance.MedianScore = scores[middle]
		if len(scores)%2 == 0 {
			result.Performance.MedianScore = round((scores[middle-1] + scores[middle]) / 2)
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	if len(ranked) > 3 {
		ranked = ranked[:3]
	}
	result.Performance.TopStudents = ranked
	result.Assignees = make([]dto.AssignmentAssignee, len(students))
	for index, assignee := range students {
		result.Assignees[index] = dto.AssignmentAssignee{
			ID: assignee.Student.ID, Name: assignee.Student.FullName(), Classroom: assignee.Student.ClassroomFor(item.AcademicYearID),
			Score: assignee.Score,
		}
		if assignee.CompletedAt != nil {
			result.Assignees[index].CompletedAt = assignee.CompletedAt.UTC().Format(timeFormat)
		}
	}
	return result
}

func examDTO(item model.Exam, detail bool, classroomFilter string) dto.Exam {
	students := item.Students
	if detail && classroomFilter != "" {
		filtered := make([]model.ExamStudent, 0, len(students))
		for _, student := range students {
			if strings.EqualFold(strings.TrimSpace(student.Student.ClassroomFor(item.AcademicYearID)), strings.TrimSpace(classroomFilter)) {
				filtered = append(filtered, student)
			}
		}
		students = filtered
	}
	marked := 0
	scoreTotal := 0.0
	performanceTotal := 0.0
	trend := make([]float64, 0, len(students))
	for _, student := range students {
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
		Class: referenceDTO(item.Class), StudentCount: len(students), MarkedCount: marked,
		AverageScore: averageScore,
	}
	if !detail {
		return result
	}
	result.Performance = &dto.Performance{
		AverageScore: averagePercent, CompletionRate: percent(marked, len(students)),
		Completed: marked, Total: len(students), Standing: standing(averagePercent, marked),
	}
	scores := append([]float64(nil), trend...)
	sort.Float64s(scores)
	if len(scores) > 0 {
		result.Performance.LowestScore = scores[0]
		result.Performance.HighestScore = scores[len(scores)-1]
		middle := len(scores) / 2
		result.Performance.MedianScore = scores[middle]
		if len(scores)%2 == 0 {
			result.Performance.MedianScore = round((scores[middle-1] + scores[middle]) / 2)
		}
	}
	ranked := make([]dto.RankedStudent, 0, len(students))
	for _, student := range students {
		if student.Score == nil || item.TotalScore <= 0 {
			continue
		}
		ranked = append(ranked, dto.RankedStudent{
			ID: student.Student.ID, Name: student.Student.FullName(),
			Score: round(*student.Score / item.TotalScore * 100),
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	if len(ranked) > 3 {
		ranked = ranked[:3]
	}
	result.Performance.TopStudents = ranked
	result.Students = make([]dto.ExamStudent, len(students))
	for index, student := range students {
		result.Students[index] = dto.ExamStudent{
			ID: student.Student.ID, Name: student.Student.FullName(), Classroom: student.Student.ClassroomFor(item.AcademicYearID),
			Score: student.Score,
		}
		if student.MarkedAt != nil {
			result.Students[index].MarkedAt = student.MarkedAt.UTC().Format(timeFormat)
		}
	}
	return result
}

func rankedStudents(students []model.Student) []dto.RankedStudent {
	ranked := make([]dto.RankedStudent, 0, len(students))
	for _, student := range students {
		performance := studentPerformance(student)
		if performance.Completed == 0 {
			continue
		}
		ranked = append(ranked, dto.RankedStudent{
			ID: student.ID, Name: student.FullName(), Score: performance.AverageScore,
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	if len(ranked) > 3 {
		ranked = ranked[:3]
	}
	return ranked
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
