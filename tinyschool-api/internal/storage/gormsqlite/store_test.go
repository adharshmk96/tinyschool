package gormsqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
	"tinyschool-api/internal/tenancy"
)

func TestStoreMigrateSeedAndPersist(t *testing.T) {
	ctx := tenancy.WithUserID(context.Background(), "usr_001")
	path := filepath.Join(t.TempDir(), "school.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AutoMigrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(ctx); err != nil {
		t.Fatal(err)
	}
	schools, total, err := store.ListSchools(ctx, storage.ListOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(schools) != 2 {
		t.Fatalf("expected two seeded schools, got total=%d length=%d", total, len(schools))
	}
	assignment, err := store.Assignment(ctx, "asg_001")
	if err != nil {
		t.Fatal(err)
	}
	if len(assignment.Students) != 3 {
		t.Fatalf("expected three assignees, got %d", len(assignment.Students))
	}
	now := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	score := 19.0
	if err := store.SetAssignmentScore(ctx, "asg_001", "stu_002", &score, &now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateStudentLog(ctx, model.StudentLog{ID: "beh_test", StudentID: "stu_001", Kind: "behaviour", Type: "incident", Note: "Test incident", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	assignment, err = store.Assignment(ctx, "asg_001")
	if err != nil {
		t.Fatal(err)
	}
	foundScore := false
	for _, assignee := range assignment.Students {
		foundScore = foundScore || (assignee.Student.ID == "stu_002" && assignee.Score != nil && *assignee.Score == score)
	}
	if !foundScore {
		t.Fatal("expected persisted assignment score")
	}
	student, err := store.Student(ctx, "stu_001")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, log := range student.Logs {
		found = found || log.ID == "beh_test"
	}
	if !found {
		t.Fatal("expected persisted student log")
	}
	if _, err := store.School(ctx, "missing"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSeedUserDataCreatesRealisticScoredData(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "school.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	ctx := context.Background()
	if err := store.AutoMigrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.db.Create(&userRecord{ID: "usr_seed", Name: "Demo Teacher", Email: "demo@example.test", PasswordHash: "unused"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := store.SeedUserData(ctx, "demo@example.test"); err != nil {
		t.Fatal(err)
	}

	var students, classes, assignments, exams, assignmentRows, scoredAssignments, examRows, scoredExams int64
	for label, query := range map[string]struct {
		target *int64
		table  string
		where  string
		args   []any
	}{
		"students":          {&students, "students", "user_id = ?", []any{"usr_seed"}},
		"classes":           {&classes, "classes", "user_id = ?", []any{"usr_seed"}},
		"assignments":       {&assignments, "assignments", "user_id = ?", []any{"usr_seed"}},
		"exams":             {&exams, "exams", "user_id = ?", []any{"usr_seed"}},
		"assignment rows":   {&assignmentRows, "assignment_students", "1 = 1", nil},
		"assignment scores": {&scoredAssignments, "assignment_students", "score IS NOT NULL", nil},
		"exam rows":         {&examRows, "exam_students", "1 = 1", nil},
		"exam scores":       {&scoredExams, "exam_students", "score IS NOT NULL", nil},
	} {
		if err := store.db.Table(query.table).Where(query.where, query.args...).Count(query.target).Error; err != nil {
			t.Fatalf("count %s: %v", label, err)
		}
	}
	if students != 40 || classes != 4 || assignments != 16 || exams != 16 {
		t.Fatalf("seeded students=%d classes=%d assignments=%d exams=%d, want 40, 4, 16, 16", students, classes, assignments, exams)
	}
	if assignmentRows != 160 || scoredAssignments != assignmentRows || examRows != 160 || scoredExams != examRows {
		t.Fatalf("assignment scores=%d/%d exam scores=%d/%d, want 160/160 for both", scoredAssignments, assignmentRows, scoredExams, examRows)
	}
}

func TestStudentGradesFollowAcademicYear(t *testing.T) {
	ctx := tenancy.WithUserID(context.Background(), "usr_001")
	store, err := Open(filepath.Join(t.TempDir(), "school.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(ctx); err != nil {
		t.Fatal(err)
	}

	student, err := store.Student(ctx, "stu_001")
	if err != nil {
		t.Fatal(err)
	}
	if got := student.GradeFor("ay_2026"); got != "Grade 7" {
		t.Fatalf("current year grade = %q, want Grade 7", got)
	}
	if got := student.GradeFor("ay_2025"); got != "Grade 6" {
		t.Fatalf("previous year grade = %q, want Grade 6", got)
	}

	// Students are not scoped to a year, so listing them is year independent.
	_, total, err := store.ListStudents(ctx, storage.ListOptions{Page: 1, PageSize: 10, AcademicYearID: "ay_2025"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 8 {
		t.Fatalf("student total = %d, want 8", total)
	}

	// Classes, however, belong to one year.
	classes, classTotal, err := store.ListClasses(ctx, storage.ListOptions{Page: 1, PageSize: 10, AcademicYearID: "ay_2025"})
	if err != nil {
		t.Fatal(err)
	}
	if classTotal != 0 || len(classes) != 0 {
		t.Fatalf("expected no classes in the previous year, got %d", classTotal)
	}

	student.Grades = []model.StudentGrade{{AcademicYearID: "ay_2026", Grade: "Grade 8"}}
	if _, err := store.UpdateStudent(ctx, student); err != nil {
		t.Fatal(err)
	}
	updated, err := store.Student(ctx, "stu_001")
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Grades) != 1 || updated.GradeFor("ay_2026") != "Grade 8" {
		t.Fatalf("grades were not replaced: %+v", updated.Grades)
	}
}

func TestClearUserDataDoesNotDeleteAnotherUsersData(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "school.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	userOne := model.User{ID: "usr_one", Name: "One", Email: "one@example.test", PasswordHash: "hash"}
	userTwo := model.User{ID: "usr_two", Name: "Two", Email: "two@example.test", PasswordHash: "hash"}
	ctxOne := tenancy.WithUserID(t.Context(), userOne.ID)
	ctxTwo := tenancy.WithUserID(t.Context(), userTwo.ID)
	if _, err := store.CreateUser(ctxOne, userOne, &model.School{ID: "sch_one", Name: "School", IsActive: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctxTwo, userTwo, &model.School{ID: "sch_two", Name: "School", IsActive: true}); err != nil {
		t.Fatal(err)
	}
	if err := store.ClearUserData(ctxOne); err != nil {
		t.Fatal(err)
	}
	if _, err := store.School(ctxOne, "sch_one"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("cleared user's school still exists: %v", err)
	}
	schools, total, err := store.ListSchools(ctxTwo, storage.ListOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(schools) != 1 || schools[0].ID != "sch_two" {
		t.Fatalf("other user's data changed: total=%d schools=%+v", total, schools)
	}
}

func TestSchoolGradesInUseAndListGradeFilter(t *testing.T) {
	ctx := tenancy.WithUserID(context.Background(), "usr_001")
	store, err := Open(filepath.Join(t.TempDir(), "school.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.AutoMigrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(ctx); err != nil {
		t.Fatal(err)
	}

	used, err := store.SchoolGradesInUse(ctx, "sch_001")
	if err != nil {
		t.Fatal(err)
	}
	if len(used) == 0 {
		t.Fatal("expected seeded grades to be in use")
	}

	classes, total, err := store.ListClasses(ctx, storage.ListOptions{
		Page: 1, PageSize: 20, AcademicYearID: "ay_2026", Grade: "Grade 7",
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(classes) != 2 {
		t.Fatalf("grade 7 classes total=%d len=%d, want 2", total, len(classes))
	}
	for _, class := range classes {
		if class.Grade != "Grade 7" {
			t.Fatalf("unexpected class grade %q", class.Grade)
		}
	}

	assignments, assignmentTotal, err := store.ListAssignments(ctx, storage.ListOptions{
		Page: 1, PageSize: 20, AcademicYearID: "ay_2026", Grade: "Grade 6",
	})
	if err != nil {
		t.Fatal(err)
	}
	if assignmentTotal < 1 || len(assignments) < 1 {
		t.Fatal("expected grade 6 assignments")
	}

	exams, examTotal, err := store.ListExams(ctx, storage.ListOptions{
		Page: 1, PageSize: 20, AcademicYearID: "ay_2026", Grade: "Grade 7",
	})
	if err != nil {
		t.Fatal(err)
	}
	if examTotal < 1 || len(exams) < 1 {
		t.Fatalf("expected grade 7 exams, got total=%d", examTotal)
	}
}
