package service

import (
	"testing"

	"tinyschool-api/internal/model"
)

func TestExamPerformanceUsesScoreSummaryNotTrend(t *testing.T) {
	first, second, third := 40.0, 70.0, 90.0
	item := model.Exam{
		TotalScore: 100,
		Students: []model.ExamStudent{
			{Student: model.Student{ID: "one", FirstName: "One"}, Score: &first},
			{Student: model.Student{ID: "two", FirstName: "Two"}, Score: &second},
			{Student: model.Student{ID: "three", FirstName: "Three"}, Score: &third},
			{Student: model.Student{ID: "four", FirstName: "Four"}},
		},
	}

	performance := examDTO(item, true).Performance
	if performance == nil {
		t.Fatal("performance is nil")
	}
	if performance.AverageScore != 66.7 || performance.MedianScore != 70 ||
		performance.LowestScore != 40 || performance.HighestScore != 90 {
		t.Fatalf("unexpected score summary: %#v", performance)
	}
	if performance.CompletionRate != 75 || len(performance.Trend) != 0 {
		t.Fatalf("completion=%d trend=%v, want 75 and no trend", performance.CompletionRate, performance.Trend)
	}
	if len(performance.TopStudents) != 3 || performance.TopStudents[0].Name != "Three" {
		t.Fatalf("unexpected top students: %#v", performance.TopStudents)
	}
}

func TestAssignmentPerformanceUsesScoreSummary(t *testing.T) {
	first, second := 15.0, 18.0
	item := model.Assignment{
		TotalScore: 20,
		Students: []model.AssignmentStudent{
			{Student: model.Student{ID: "one", FirstName: "One"}, Score: &first},
			{Student: model.Student{ID: "two", FirstName: "Two"}, Score: &second},
			{Student: model.Student{ID: "three", FirstName: "Three"}},
		},
	}

	performance := assignmentDTO(item, true).Performance
	if performance == nil {
		t.Fatal("performance is nil")
	}
	if performance.AverageScore != 82.5 || performance.MedianScore != 82.5 ||
		performance.LowestScore != 75 || performance.HighestScore != 90 {
		t.Fatalf("unexpected score summary: %#v", performance)
	}
	if performance.CompletionRate != 67 || len(performance.TopStudents) != 2 ||
		performance.TopStudents[0].Name != "Two" {
		t.Fatalf("unexpected completion or ranking: %#v", performance)
	}
}

func TestClassPerformanceSeparatesCompletionAndGroupsExamMonths(t *testing.T) {
	assignmentScore, januaryScore, februaryScore := 8.0, 75.0, 90.0
	students := []model.Student{
		{
			ID: "one", FirstName: "One",
			Assignments: []model.Result{
				{ID: "a1", DueAt: "2026-01-05", Score: &assignmentScore, TotalScore: 10},
				{ID: "a2", DueAt: "2026-01-12", TotalScore: 10},
			},
			Exams: []model.Result{
				{ID: "e1", DueAt: "2026-01-20", Score: &januaryScore, TotalScore: 100},
				{ID: "e2", DueAt: "2026-02-20", Score: &februaryScore, TotalScore: 100},
			},
		},
	}

	performance := classPerformance(students)
	if performance.AssignmentCompletion != 50 || performance.ExamCompletion != 100 {
		t.Fatalf("assignment=%d exam=%d, want 50 and 100", performance.AssignmentCompletion, performance.ExamCompletion)
	}
	if len(performance.MonthlyExamTrend) != 2 ||
		performance.MonthlyExamTrend[0].Label != "Jan 2026" ||
		performance.MonthlyExamTrend[1].Value != 90 {
		t.Fatalf("unexpected monthly trend: %#v", performance.MonthlyExamTrend)
	}
	if len(performance.TopStudents) != 1 || performance.TopStudents[0].Name != "One" {
		t.Fatalf("unexpected top students: %#v", performance.TopStudents)
	}
}

func TestStudentPerformanceSeparatesMonthlyAssignmentAndExamAnalytics(t *testing.T) {
	assignmentOne, assignmentTwo, exam := 8.0, 6.0, 90.0
	student := model.Student{
		Assignments: []model.Result{
			{ID: "a1", Name: "Reading", DueAt: "2026-01-05", Score: &assignmentOne, TotalScore: 10},
			{ID: "a2", Name: "Writing", DueAt: "2026-02-05", Score: &assignmentTwo, TotalScore: 10},
			{ID: "a3", Name: "Pending", DueAt: "2026-02-10", TotalScore: 10},
		},
		Exams: []model.Result{
			{ID: "e1", Name: "Midterm", DueAt: "2026-02-20", Score: &exam, TotalScore: 100},
		},
	}

	performance := studentPerformance(student)
	if performance.AssignmentAverage != 70 || performance.ExamAverage != 90 {
		t.Fatalf("assignment=%v exam=%v, want 70 and 90", performance.AssignmentAverage, performance.ExamAverage)
	}
	if performance.AssignmentCompletion != 67 || performance.ExamCompletion != 100 {
		t.Fatalf("assignment completion=%d exam completion=%d", performance.AssignmentCompletion, performance.ExamCompletion)
	}
	if performance.StrongestArea != "Exams" || performance.ScoreConsistency != 88 {
		t.Fatalf("strongest=%q consistency=%d", performance.StrongestArea, performance.ScoreConsistency)
	}
	if performance.BestResult == nil || performance.BestResult.Name != "Midterm" {
		t.Fatalf("unexpected best result: %#v", performance.BestResult)
	}
	if len(performance.MonthlyAssignmentTrend) != 2 ||
		performance.MonthlyAssignmentTrend[0].Label != "Jan 2026" ||
		len(performance.MonthlyExamTrend) != 1 ||
		performance.MonthlyExamTrend[0].Value != 90 {
		t.Fatalf("assignment trend=%#v exam trend=%#v", performance.MonthlyAssignmentTrend, performance.MonthlyExamTrend)
	}
}
