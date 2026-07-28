package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

type fakeStorage struct {
	storage.Storage
	user        model.User
	session     model.Session
	school      model.School
	year        model.AcademicYear
	assignment  model.Assignment
	student     model.Student
	createdYear model.AcademicYear
	scoreWasSet bool
	deleted     bool
}

func (f *fakeStorage) School(_ context.Context, id string) (model.School, error) {
	if f.school.ID != id {
		return model.School{}, storage.ErrNotFound
	}
	return f.school, nil
}

func (f *fakeStorage) CreateAcademicYear(_ context.Context, value model.AcademicYear) (model.AcademicYear, error) {
	f.createdYear = value
	return value, nil
}

func (f *fakeStorage) AcademicYear(_ context.Context, _ string) (model.AcademicYear, error) {
	if f.year.ID == "" {
		return model.AcademicYear{}, storage.ErrNotFound
	}
	return f.year, nil
}

func (f *fakeStorage) Assignment(_ context.Context, _ string) (model.Assignment, error) {
	if f.assignment.ID == "" {
		return model.Assignment{}, storage.ErrNotFound
	}
	return f.assignment, nil
}

func (f *fakeStorage) SetAssignmentScore(_ context.Context, _ string, _ string, _ *float64, _ *time.Time) error {
	f.scoreWasSet = true
	return nil
}

func (f *fakeStorage) DeleteAssignment(_ context.Context, _ string) error {
	f.deleted = true
	return nil
}

func (f *fakeStorage) Student(_ context.Context, _ string) (model.Student, error) {
	if f.student.ID == "" {
		return model.Student{}, storage.ErrNotFound
	}
	return f.student, nil
}

func (f *fakeStorage) DeleteStudent(_ context.Context, _ string) error {
	f.deleted = true
	return nil
}

func (f *fakeStorage) CreateUser(_ context.Context, value model.User, _ *model.School) (model.User, error) {
	f.user = value
	return value, nil
}

func (f *fakeStorage) CreateSession(_ context.Context, value model.Session) (model.Session, error) {
	f.session = value
	return value, nil
}

func (f *fakeStorage) Session(_ context.Context, id string) (model.Session, error) {
	if f.session.ID != id {
		return model.Session{}, storage.ErrNotFound
	}
	return f.session, nil
}

func (f *fakeStorage) UserByID(_ context.Context, id string) (model.User, error) {
	if f.user.ID != id {
		return model.User{}, storage.ErrNotFound
	}
	return f.user, nil
}

func TestCreateAcademicYearCalculatesConsecutiveDates(t *testing.T) {
	store := &fakeStorage{school: model.School{ID: "sch_1", Grades: []string{"Grade 1"}}}
	app := New(store, WithIDGenerator(func(prefix string) (string, error) {
		return prefix + "_test", nil
	}))

	result, err := app.CreateAcademicYear(context.Background(), dto.AcademicYearRequest{
		SchoolID: "sch_1", Name: "2026", StartDate: "2026-06-01", IsCurrent: true,
		Segments: []dto.SegmentRequest{
			{Name: "Term 1", Type: "term", DurationDays: 2},
			{Name: "Break", Type: "vacation", DurationDays: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.DurationDays != 3 || result.EndDate != "2026-06-03" {
		t.Fatalf("unexpected year calculation: %+v", result)
	}
	if result.Segments[0].EndDate != "2026-06-02" || result.Segments[1].StartDate != "2026-06-03" {
		t.Fatalf("segments are not consecutive: %+v", result.Segments)
	}
}

func TestCreateAcademicYearRejectsInvalidSegment(t *testing.T) {
	app := New(&fakeStorage{})
	_, err := app.CreateAcademicYear(context.Background(), dto.AcademicYearRequest{
		SchoolID: "sch_1", Name: "2026", StartDate: "2026-06-01",
		Segments: []dto.SegmentRequest{{Name: "Break", Type: "holiday", DurationDays: 1}},
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestAssignmentScoreRejectsOutOfRangeBeforeWrite(t *testing.T) {
	store := &fakeStorage{assignment: model.Assignment{ID: "asg_1", TotalScore: 20}}
	app := New(store)
	_, err := app.SetAssignmentScore(context.Background(), "asg_1", "stu_1", dto.ScoreRequest{Score: 21})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if store.scoreWasSet {
		t.Fatal("storage write occurred for invalid score")
	}
}

func TestDeleteAssignmentRejectsRecordedScore(t *testing.T) {
	score := 18.0
	store := &fakeStorage{assignment: model.Assignment{
		ID: "asg_1",
		Students: []model.AssignmentStudent{{
			Student: model.Student{ID: "stu_1"},
			Score:   &score,
		}},
	}}
	err := New(store).DeleteAssignment(context.Background(), "asg_1")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if store.deleted {
		t.Fatal("storage delete occurred for a scored assignment")
	}
}

func TestDeleteStudentListsConflictingData(t *testing.T) {
	store := &fakeStorage{student: model.Student{
		ID:      "stu_1",
		Classes: []model.Reference{{ID: "cls_1", Name: "Grade 8"}},
		Assignments: []model.Result{
			{ID: "asg_1", Name: "Math homework"},
			{ID: "asg_2", Name: "Reading"},
		},
		Logs: []model.StudentLog{{ID: "note_1", Kind: "note"}},
	}}

	err := New(store).DeleteStudent(context.Background(), "stu_1")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	want := `student cannot be deleted because it has related data: class "Grade 8"; assignments "Math homework", "Reading"; 1 note`
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
	if store.deleted {
		t.Fatal("storage delete occurred for a student with related data")
	}
}

func TestRefreshAllowsExpiredTokenWhileSessionActive(t *testing.T) {
	store := &fakeStorage{}
	now := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
	app := New(
		store,
		WithClock(func() time.Time { return now }),
		WithJWTSecret([]byte("01234567890123456789012345678901")),
		WithSessionDuration(time.Hour),
		WithTokenDuration(time.Minute),
		WithIDGenerator(func(prefix string) (string, error) { return prefix + "_1", nil }),
	)
	registered, err := app.Register(context.Background(), dto.RegisterRequest{
		Name: "Alex", Email: "alex@example.test", Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	now = now.Add(2 * time.Minute)
	if _, err := app.Authenticate(context.Background(), registered.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected expired token, got %v", err)
	}
	refreshed, err := app.Refresh(context.Background(), registered.Token)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.Token == registered.Token {
		t.Fatal("expected a rotated token")
	}
	if _, err := app.Authenticate(context.Background(), refreshed.Token); err != nil {
		t.Fatalf("new token was not accepted: %v", err)
	}
}
