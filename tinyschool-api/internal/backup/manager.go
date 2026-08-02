package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

var (
	ErrBusy     = errors.New("backup operation is already running")
	ErrNotFound = errors.New("backup not found")
)

var managedName = regexp.MustCompile(`^tinyschool-(\d{8}T\d{6}Z)(?:-pre-restore)?(?:-\d+)?\.db$`)

type Settings struct {
	Enabled    bool       `json:"enabled"`
	Frequency  string     `json:"frequency"`
	RunAt      string     `json:"runAt"`
	MaxBackups int        `json:"maxBackups"`
	NextRunAt  *time.Time `json:"nextRunAt,omitempty"`
}

type File struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

type Manager struct {
	databasePath string
	directory    string
	logger       *slog.Logger
	now          func() time.Time
	operation    chan struct{}
	reschedule   chan struct{}
	mu           sync.RWMutex
	nextRunAt    *time.Time
}

func New(databasePath string, logger *slog.Logger) (*Manager, error) {
	if logger == nil {
		logger = slog.Default()
	}
	m := &Manager{
		databasePath: databasePath,
		directory:    filepath.Join(filepath.Dir(databasePath), "backups"),
		logger:       logger,
		now:          time.Now,
		operation:    make(chan struct{}, 1),
		reschedule:   make(chan struct{}, 1),
	}
	if err := os.MkdirAll(m.directory, 0o700); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}
	db, err := open(databasePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS backup_settings (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		enabled INTEGER NOT NULL DEFAULT 0,
		frequency TEXT NOT NULL DEFAULT 'daily',
		run_at TEXT NOT NULL DEFAULT '02:00',
		max_backups INTEGER NOT NULL DEFAULT 14
	)`); err != nil {
		return nil, fmt.Errorf("create backup settings: %w", err)
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO backup_settings (id) VALUES (1)`); err != nil {
		return nil, fmt.Errorf("initialize backup settings: %w", err)
	}
	return m, nil
}

func (m *Manager) Start(ctx context.Context) {
	go m.schedule(ctx)
}

func (m *Manager) Settings(ctx context.Context) (Settings, error) {
	db, err := open(m.databasePath)
	if err != nil {
		return Settings{}, err
	}
	defer db.Close()
	var value Settings
	if err := db.QueryRowContext(ctx, `SELECT enabled, frequency, run_at, max_backups FROM backup_settings WHERE id = 1`).Scan(
		&value.Enabled, &value.Frequency, &value.RunAt, &value.MaxBackups,
	); err != nil {
		return Settings{}, fmt.Errorf("read backup settings: %w", err)
	}
	m.mu.RLock()
	if m.nextRunAt != nil {
		next := *m.nextRunAt
		value.NextRunAt = &next
	}
	m.mu.RUnlock()
	return value, nil
}

func (m *Manager) SaveSettings(ctx context.Context, value Settings) (Settings, error) {
	if err := validateSettings(value); err != nil {
		return Settings{}, err
	}
	db, err := open(m.databasePath)
	if err != nil {
		return Settings{}, err
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `UPDATE backup_settings SET enabled = ?, frequency = ?, run_at = ?, max_backups = ? WHERE id = 1`,
		value.Enabled, value.Frequency, value.RunAt, value.MaxBackups,
	); err != nil {
		return Settings{}, fmt.Errorf("save backup settings: %w", err)
	}
	select {
	case m.reschedule <- struct{}{}:
	default:
	}
	return m.Settings(ctx)
}

func validateSettings(value Settings) error {
	switch value.Frequency {
	case "daily", "every_2_days", "weekly":
	default:
		return fmt.Errorf("frequency must be daily, every_2_days, or weekly")
	}
	if _, err := time.Parse("15:04", value.RunAt); err != nil {
		return fmt.Errorf("runAt must use HH:mm format")
	}
	if value.MaxBackups < 1 || value.MaxBackups > 100 {
		return fmt.Errorf("maxBackups must be between 1 and 100")
	}
	return nil
}

func (m *Manager) List() ([]File, error) {
	entries, err := os.ReadDir(m.directory)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	files := make([]File, 0, len(entries))
	for _, entry := range entries {
		created, ok := timeFromName(entry.Name())
		if entry.IsDir() || !ok || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("inspect backup %s: %w", entry.Name(), err)
		}
		files = append(files, File{Name: entry.Name(), Size: info.Size(), CreatedAt: created})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].CreatedAt.After(files[j].CreatedAt) })
	return files, nil
}

func (m *Manager) Create(ctx context.Context) (File, error) {
	if !m.acquire() {
		return File{}, ErrBusy
	}
	defer m.release()
	return m.create(ctx, false)
}

func (m *Manager) create(ctx context.Context, safety bool) (File, error) {
	started := m.now().UTC()
	suffix := ""
	if safety {
		suffix = "-pre-restore"
	}
	name := "tinyschool-" + started.Format("20060102T150405Z") + suffix + ".db"
	path := filepath.Join(m.directory, name)
	for sequence := 2; ; sequence++ {
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			break
		} else if err != nil {
			return File{}, fmt.Errorf("inspect backup destination: %w", err)
		}
		name = "tinyschool-" + started.Format("20060102T150405Z") + suffix + fmt.Sprintf("-%d.db", sequence)
		path = filepath.Join(m.directory, name)
	}
	temporary, err := os.CreateTemp(m.directory, ".backup-*.tmp")
	if err != nil {
		return File{}, fmt.Errorf("create backup temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return File{}, fmt.Errorf("close backup temporary file: %w", err)
	}
	_ = os.Remove(temporaryPath)
	defer os.Remove(temporaryPath)

	m.logger.Info("database backup started", "name", name)
	if err := copyDatabase(ctx, temporaryPath, m.databasePath); err != nil {
		return File{}, fmt.Errorf("create database backup: %w", err)
	}
	if err := validateDatabase(ctx, temporaryPath); err != nil {
		return File{}, err
	}
	if err := os.Chmod(temporaryPath, 0o600); err != nil {
		return File{}, fmt.Errorf("secure backup file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return File{}, fmt.Errorf("publish backup: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return File{}, fmt.Errorf("inspect created backup: %w", err)
	}
	file := File{Name: name, Size: info.Size(), CreatedAt: started}
	if !safety {
		settings, err := m.Settings(ctx)
		if err != nil {
			return File{}, err
		}
		if err := m.prune(settings.MaxBackups); err != nil {
			return File{}, err
		}
	}
	m.logger.Info("database backup completed", "name", name, "size", file.Size, "duration", time.Since(started))
	return file, nil
}

func (m *Manager) Path(name string) (string, File, error) {
	files, err := m.List()
	if err != nil {
		return "", File{}, err
	}
	for _, file := range files {
		if file.Name == name {
			return filepath.Join(m.directory, file.Name), file, nil
		}
	}
	return "", File{}, ErrNotFound
}

func (m *Manager) Restore(ctx context.Context, name string) (File, error) {
	if !m.acquire() {
		return File{}, ErrBusy
	}
	defer m.release()
	path, selected, err := m.Path(name)
	if err != nil {
		return File{}, err
	}
	if err := validateDatabase(ctx, path); err != nil {
		return File{}, err
	}
	if _, err := m.create(ctx, true); err != nil {
		return File{}, fmt.Errorf("create pre-restore backup: %w", err)
	}
	m.logger.Warn("database restore started", "name", name)
	if err := copyDatabase(ctx, m.databasePath, path); err != nil {
		return File{}, fmt.Errorf("restore database: %w", err)
	}
	if err := validateDatabase(ctx, m.databasePath); err != nil {
		return File{}, fmt.Errorf("validate restored database: %w", err)
	}
	select {
	case m.reschedule <- struct{}{}:
	default:
	}
	m.logger.Warn("database restore completed", "name", name)
	return selected, nil
}

func (m *Manager) acquire() bool {
	select {
	case m.operation <- struct{}{}:
		return true
	default:
		return false
	}
}

func (m *Manager) release() { <-m.operation }

func (m *Manager) prune(max int) error {
	files, err := m.List()
	if err != nil {
		return err
	}
	if len(files) <= max {
		return nil
	}
	for _, file := range files[max:] {
		if err := os.Remove(filepath.Join(m.directory, file.Name)); err != nil {
			return fmt.Errorf("remove old backup %s: %w", file.Name, err)
		}
	}
	return nil
}

func (m *Manager) schedule(ctx context.Context) {
	for {
		settings, err := m.Settings(ctx)
		if err != nil {
			m.logger.Error("read scheduled backup settings", "error", err)
			if !wait(ctx, m.reschedule, time.Minute) {
				return
			}
			continue
		}
		if !settings.Enabled {
			m.setNext(nil)
			select {
			case <-ctx.Done():
				return
			case <-m.reschedule:
				continue
			}
		}
		next := nextRun(m.now(), settings)
		m.setNext(&next)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-m.reschedule:
			timer.Stop()
			continue
		case <-timer.C:
			if _, err := m.Create(ctx); err != nil {
				m.logger.Error("scheduled database backup failed", "error", err)
			}
		}
	}
}

func (m *Manager) setNext(next *time.Time) {
	m.mu.Lock()
	m.nextRunAt = next
	m.mu.Unlock()
}

func nextRun(now time.Time, settings Settings) time.Time {
	parsed, _ := time.Parse("15:04", settings.RunAt)
	candidate := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	interval := 24 * time.Hour
	if settings.Frequency == "every_2_days" {
		interval = 48 * time.Hour
	} else if settings.Frequency == "weekly" {
		interval = 7 * 24 * time.Hour
	}
	if !candidate.After(now) {
		candidate = candidate.Add(interval)
	}
	return candidate
}

func wait(ctx context.Context, reschedule <-chan struct{}, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-reschedule:
		return true
	case <-timer.C:
		return true
	}
}

func timeFromName(name string) (time.Time, bool) {
	matches := managedName.FindStringSubmatch(name)
	if len(matches) != 2 {
		return time.Time{}, false
	}
	value, err := time.Parse("20060102T150405Z", matches[1])
	return value, err == nil
}

func open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(path)+"?_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func copyDatabase(ctx context.Context, destinationPath, sourcePath string) error {
	destination, err := open(destinationPath)
	if err != nil {
		return err
	}
	defer destination.Close()
	source, err := open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	destinationConn, err := destination.Conn(ctx)
	if err != nil {
		return err
	}
	defer destinationConn.Close()
	sourceConn, err := source.Conn(ctx)
	if err != nil {
		return err
	}
	defer sourceConn.Close()
	return destinationConn.Raw(func(destinationDriver any) error {
		dest, ok := destinationDriver.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("unexpected destination sqlite driver")
		}
		return sourceConn.Raw(func(sourceDriver any) error {
			src, ok := sourceDriver.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("unexpected source sqlite driver")
			}
			backup, err := dest.Backup("main", src, "main")
			if err != nil {
				return err
			}
			for {
				done, err := backup.Step(128)
				if err != nil {
					_ = backup.Close()
					return err
				}
				if done {
					return backup.Close()
				}
				select {
				case <-ctx.Done():
					_ = backup.Close()
					return ctx.Err()
				case <-time.After(10 * time.Millisecond):
				}
			}
		})
	})
}

func validateDatabase(ctx context.Context, path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect backup: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("backup is not a regular file")
	}
	db, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(path)+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open backup for validation: %w", err)
	}
	defer db.Close()
	var integrity string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		return fmt.Errorf("check backup integrity: %w", err)
	}
	if strings.ToLower(integrity) != "ok" {
		return fmt.Errorf("backup integrity check failed: %s", integrity)
	}
	var users int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'users'`).Scan(&users); err != nil {
		return fmt.Errorf("check backup schema: %w", err)
	}
	if users != 1 {
		return fmt.Errorf("backup does not contain the Tiny School schema")
	}
	return nil
}
