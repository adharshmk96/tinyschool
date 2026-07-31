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
		&userRecord{}, &sessionRecord{}, &passwordResetRecord{}, &schoolRecord{}, &schoolClassroomRecord{},
		&academicYearRecord{}, &academicSegmentRecord{}, &studentRecord{},
		&studentClassroomRecord{}, &classRecord{}, &classClassroomRecord{}, &classStudentRecord{}, &studentLogRecord{},
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
	if err := s.backfillStudentGrades(ctx); err != nil {
		return err
	}
	if err := s.backfillClassrooms(ctx); err != nil {
		return err
	}
	return nil
}

// backfillStudentGrades moves the legacy students.grade column into the
// per-academic-year student_grades table (then classrooms backfill copies it).
func (s *Store) backfillStudentGrades(ctx context.Context) error {
	db := s.db.WithContext(ctx)
	var legacyColumns int64
	if err := db.Raw("SELECT COUNT(*) FROM pragma_table_info('students') WHERE name = 'grade'").Scan(&legacyColumns).Error; err != nil {
		return fmt.Errorf("inspect students table: %w", err)
	}
	if legacyColumns == 0 {
		return nil
	}
	var hasStudentGrades int64
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'student_grades'").Scan(&hasStudentGrades).Error; err != nil {
		return fmt.Errorf("inspect student_grades table: %w", err)
	}
	if hasStudentGrades == 0 {
		return nil
	}
	err := db.Exec(`
		INSERT INTO student_grades (student_id, academic_year_id, grade)
		SELECT s.id, ay.id, s.grade
		FROM students s
		JOIN academic_years ay ON ay.school_id = s.school_id AND ay.is_current = 1
		WHERE s.grade IS NOT NULL AND s.grade <> ''
		  AND NOT EXISTS (SELECT 1 FROM student_grades sg WHERE sg.student_id = s.id AND sg.academic_year_id = ay.id)
	`).Error
	if err != nil {
		return fmt.Errorf("backfill student grades: %w", err)
	}
	return nil
}

// backfillClassrooms copies legacy grade tables/columns into classroom tables.
func (s *Store) backfillClassrooms(ctx context.Context) error {
	db := s.db.WithContext(ctx)
	if err := copyTableIfExists(db, "school_grades", `
		INSERT INTO school_classrooms (school_id, classroom, position)
		SELECT school_id, grade, position FROM school_grades
		WHERE NOT EXISTS (
			SELECT 1 FROM school_classrooms sc
			WHERE sc.school_id = school_grades.school_id AND LOWER(sc.classroom) = LOWER(school_grades.grade)
		)
	`); err != nil {
		return fmt.Errorf("backfill school classrooms: %w", err)
	}
	if err := copyTableIfExists(db, "student_grades", `
		INSERT INTO student_classrooms (student_id, academic_year_id, classroom)
		SELECT student_id, academic_year_id, grade FROM student_grades
		WHERE NOT EXISTS (
			SELECT 1 FROM student_classrooms sc
			WHERE sc.student_id = student_grades.student_id AND sc.academic_year_id = student_grades.academic_year_id
		)
	`); err != nil {
		return fmt.Errorf("backfill student classrooms: %w", err)
	}
	var hasGradeColumn int64
	if err := db.Raw("SELECT COUNT(*) FROM pragma_table_info('classes') WHERE name = 'grade'").Scan(&hasGradeColumn).Error; err != nil {
		return fmt.Errorf("inspect classes.grade: %w", err)
	}
	if hasGradeColumn > 0 {
		if err := db.Exec(`
			INSERT INTO class_classrooms (class_id, classroom)
			SELECT id, grade FROM classes
			WHERE grade IS NOT NULL AND grade <> ''
			  AND NOT EXISTS (
				SELECT 1 FROM class_classrooms cc
				WHERE cc.class_id = classes.id AND LOWER(cc.classroom) = LOWER(classes.grade)
			  )
		`).Error; err != nil {
			return fmt.Errorf("backfill class classrooms: %w", err)
		}
	}
	return nil
}

func copyTableIfExists(db *gorm.DB, table, insertSQL string) error {
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	return db.Exec(insertSQL).Error
}

func userID(ctx context.Context) string { return tenancy.UserID(ctx) }

func (s *Store) ClearUserData(ctx context.Context) error {
	id := userID(ctx)
	if id == "" {
		return fmt.Errorf("clear user data: missing user identity")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return clearOwnedData(tx, id)
	})
}

// DeleteUser removes the account together with everything it owns. The back
// office is the only caller, so the owner id is passed explicitly instead of
// being read from the tenancy context.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("delete user: missing user identity")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := clearOwnedData(tx, id); err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM password_reset_tokens WHERE user_id = ?", id).Error; err != nil {
			return err
		}
		result := tx.Exec("DELETE FROM users WHERE id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return storage.ErrNotFound
		}
		return nil
	})
}

func clearOwnedData(tx *gorm.DB, id string) error {
	steps := []struct {
		query string
		args  []any
	}{
		{"DELETE FROM assignment_students WHERE assignment_id IN (SELECT id FROM assignments WHERE user_id = ?)", []any{id}},
		{"DELETE FROM exam_students WHERE exam_id IN (SELECT id FROM exams WHERE user_id = ?)", []any{id}},
		{"DELETE FROM student_logs WHERE student_id IN (SELECT id FROM students WHERE user_id = ?)", []any{id}},
		{"DELETE FROM student_classrooms WHERE student_id IN (SELECT id FROM students WHERE user_id = ?)", []any{id}},
		{"DELETE FROM class_classrooms WHERE class_id IN (SELECT id FROM classes WHERE user_id = ?)", []any{id}},
		{"DELETE FROM class_students WHERE class_id IN (SELECT id FROM classes WHERE user_id = ?)", []any{id}},
		{"DELETE FROM exams WHERE user_id = ?", []any{id}},
		{"DELETE FROM assignments WHERE user_id = ?", []any{id}},
		{"DELETE FROM classes WHERE user_id = ?", []any{id}},
		{"DELETE FROM students WHERE user_id = ?", []any{id}},
		{"DELETE FROM academic_segments WHERE academic_year_id IN (SELECT id FROM academic_years WHERE user_id = ?)", []any{id}},
		{"DELETE FROM academic_years WHERE user_id = ?", []any{id}},
		{"DELETE FROM school_classrooms WHERE school_id IN (SELECT id FROM schools WHERE user_id = ?)", []any{id}},
		{"DELETE FROM schools WHERE user_id = ?", []any{id}},
	}
	for _, step := range steps {
		if err := tx.Exec(step.query, step.args...).Error; err != nil {
			return err
		}
	}
	return nil
}
