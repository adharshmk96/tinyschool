package gormsqlite

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tinyschool-api/internal/storage"
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
	if err := s.db.WithContext(ctx).Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_academic_year_current_school ON academic_years(school_id) WHERE is_current = 1",
	).Error; err != nil {
		return fmt.Errorf("create current academic year index: %w", err)
	}
	return nil
}
