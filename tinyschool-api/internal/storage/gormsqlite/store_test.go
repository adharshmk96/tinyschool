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
