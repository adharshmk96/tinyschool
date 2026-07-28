package gormsqlite

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tinyschool-api/internal/storage"
	"tinyschool-api/internal/tenancy"
)

type Store struct {
	db  *gorm.DB
	sql *sql.DB
}

var _ storage.Storage = (*Store)(nil)

func Open(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: false,
		TranslateError:                           true,
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sqlite connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	store := &Store{db: db, sql: sqlDB}
	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	} {
		if err := db.Exec(pragma).Error; err != nil {
			_ = sqlDB.Close()
			return nil, fmt.Errorf("%s: %w", pragma, err)
		}
	}
	if err := store.Ping(context.Background()); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Ping(ctx context.Context) error {
	if err := s.sql.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sqlite: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if err := s.sql.Close(); err != nil {
		return fmt.Errorf("close sqlite: %w", err)
	}
	return nil
}

func (s *Store) AutoMigrate(ctx context.Context) error {
	err := s.db.WithContext(ctx).AutoMigrate(
		&userRecord{}, &sessionRecord{}, &schoolRecord{}, &schoolGradeRecord{},
		&academicYearRecord{}, &academicSegmentRecord{}, &studentRecord{},
		&classRecord{}, &classStudentRecord{}, &studentLogRecord{},
		&assignmentRecord{}, &assignmentStudentRecord{},
		&examRecord{}, &examStudentRecord{},
	)
	if err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	// Older single-user databases did not store ownership. Assign their data to
	// the first account once, before enforcing user-scoped access.
	var firstUser userRecord
	if err := s.db.WithContext(ctx).Order("id").First(&firstUser).Error; err == nil {
		for _, table := range []string{"schools", "academic_years", "students", "classes", "assignments", "exams"} {
			if err := s.db.WithContext(ctx).Table(table).Where("user_id = '' OR user_id IS NULL").Update("user_id", firstUser.ID).Error; err != nil {
				return fmt.Errorf("backfill %s ownership: %w", table, err)
			}
		}
	}
	_ = s.db.WithContext(ctx).Exec("DROP INDEX IF EXISTS idx_schools_name").Error
	if err := s.db.WithContext(ctx).Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_academic_year_current_school ON academic_years(school_id) WHERE is_current = 1",
	).Error; err != nil {
		return fmt.Errorf("create current academic year index: %w", err)
	}
	return nil
}

func userID(ctx context.Context) string { return tenancy.UserID(ctx) }

func (s *Store) ClearUserData(ctx context.Context) error {
	id := userID(ctx)
	if id == "" {
		return fmt.Errorf("clear user data: missing user identity")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		steps := []struct {
			query string
			args  []any
		}{
			{"DELETE FROM assignment_students WHERE assignment_id IN (SELECT id FROM assignments WHERE user_id = ?)", []any{id}},
			{"DELETE FROM exam_students WHERE exam_id IN (SELECT id FROM exams WHERE user_id = ?)", []any{id}},
			{"DELETE FROM student_logs WHERE student_id IN (SELECT id FROM students WHERE user_id = ?)", []any{id}},
			{"DELETE FROM class_students WHERE class_id IN (SELECT id FROM classes WHERE user_id = ?)", []any{id}},
			{"DELETE FROM exams WHERE user_id = ?", []any{id}},
			{"DELETE FROM assignments WHERE user_id = ?", []any{id}},
			{"DELETE FROM classes WHERE user_id = ?", []any{id}},
			{"DELETE FROM students WHERE user_id = ?", []any{id}},
			{"DELETE FROM academic_segments WHERE academic_year_id IN (SELECT id FROM academic_years WHERE user_id = ?)", []any{id}},
			{"DELETE FROM academic_years WHERE user_id = ?", []any{id}},
			{"DELETE FROM school_grades WHERE school_id IN (SELECT id FROM schools WHERE user_id = ?)", []any{id}},
			{"DELETE FROM schools WHERE user_id = ?", []any{id}},
		}
		for _, step := range steps {
			if err := tx.Exec(step.query, step.args...).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
