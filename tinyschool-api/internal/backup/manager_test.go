package backup

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"
)

func newTestManager(t *testing.T) (*Manager, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tinyschool.db")
	db, err := open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, name TEXT NOT NULL); INSERT INTO users VALUES ('one', 'Before')`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	manager, err := New(path, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return manager, path
}

func TestCreateRestoreAndRetention(t *testing.T) {
	manager, path := newTestManager(t)
	clock := time.Date(2026, 8, 2, 2, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return clock }
	if _, err := manager.SaveSettings(t.Context(), Settings{Frequency: "daily", RunAt: "02:00", MaxBackups: 1}); err != nil {
		t.Fatal(err)
	}
	first, err := manager.Create(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	db, err := open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE users SET name = 'After' WHERE id = 'one'`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(time.Hour)
	if _, err := manager.Create(t.Context()); err != nil {
		t.Fatal(err)
	}
	files, err := manager.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name == first.Name {
		t.Fatalf("retained files = %#v", files)
	}

	// Keep more files, create a snapshot of "Before", then prove restore writes
	// that snapshot back into the live database.
	if _, err := manager.SaveSettings(t.Context(), Settings{Frequency: "daily", RunAt: "02:00", MaxBackups: 10}); err != nil {
		t.Fatal(err)
	}
	db, _ = open(path)
	_, _ = db.Exec(`UPDATE users SET name = 'Before' WHERE id = 'one'`)
	_ = db.Close()
	clock = clock.Add(time.Hour)
	restorePoint, err := manager.Create(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	db, _ = open(path)
	_, _ = db.Exec(`UPDATE users SET name = 'Changed' WHERE id = 'one'`)
	_ = db.Close()
	clock = clock.Add(time.Hour)
	if _, err := manager.Restore(context.Background(), restorePoint.Name); err != nil {
		t.Fatal(err)
	}
	db, _ = open(path)
	defer db.Close()
	var name string
	if err := db.QueryRow(`SELECT name FROM users WHERE id = 'one'`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Before" {
		t.Fatalf("restored name = %q", name)
	}
	files, _ = manager.List()
	if len(files) < 3 {
		t.Fatalf("expected restore point and safety backup, got %#v", files)
	}
}

func TestSettingsValidationAndPathSafety(t *testing.T) {
	manager, _ := newTestManager(t)
	if _, err := manager.SaveSettings(t.Context(), Settings{Frequency: "hourly", RunAt: "02:00", MaxBackups: 5}); err == nil {
		t.Fatal("expected invalid frequency")
	}
	if _, _, err := manager.Path("../tinyschool.db"); err != ErrNotFound {
		t.Fatalf("path traversal error = %v", err)
	}
}

func TestNextRun(t *testing.T) {
	now := time.Date(2026, 8, 2, 3, 0, 0, 0, time.FixedZone("local", 7*60*60))
	next := nextRun(now, Settings{Frequency: "daily", RunAt: "02:00"})
	if want := now.Add(23 * time.Hour); !next.Equal(want) {
		t.Fatalf("next = %s, want %s", next, want)
	}
}
