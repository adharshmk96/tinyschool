package gormsqlite

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
)

func (s *Store) CurrentUser(ctx context.Context) (model.User, error) {
	var record userRecord
	err := s.db.WithContext(ctx).Order("id").First(&record).Error
	return userModel(record), storageError(err)
}

func (s *Store) UserByID(ctx context.Context, id string) (model.User, error) {
	var record userRecord
	err := s.db.WithContext(ctx).First(&record, "id = ?", id).Error
	return userModel(record), storageError(err)
}

func (s *Store) UserByEmail(ctx context.Context, email string) (model.User, error) {
	var record userRecord
	err := s.db.WithContext(ctx).First(&record, "email = ? COLLATE NOCASE", email).Error
	return userModel(record), storageError(err)
}

func (s *Store) CreateUser(ctx context.Context, user model.User, school *model.School) (model.User, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&userRecord{ID: user.ID, Name: user.Name, Email: user.Email, PasswordHash: user.PasswordHash}).Error; err != nil {
			return err
		}
		if school != nil {
			return createSchool(tx, user.ID, *school)
		}
		return nil
	})
	return user, storageError(err)
}

func (s *Store) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	result := s.db.WithContext(ctx).Model(&userRecord{}).Where("id = ?", user.ID).
		Updates(map[string]any{"name": user.Name, "email": user.Email, "password_hash": user.PasswordHash})
	if result.Error != nil {
		return model.User{}, storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return model.User{}, storageError(gorm.ErrRecordNotFound)
	}
	return s.UserByID(ctx, user.ID)
}

func (s *Store) CreateSession(ctx context.Context, session model.Session) (model.Session, error) {
	record := sessionRecord{ID: session.ID, UserID: session.UserID, ExpiresAt: session.ExpiresAt, RevokedAt: session.RevokedAt, CreatedAt: session.CreatedAt}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return model.Session{}, storageError(err)
	}
	return session, nil
}

func (s *Store) Session(ctx context.Context, id string) (model.Session, error) {
	var record sessionRecord
	if err := s.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return model.Session{}, storageError(err)
	}
	return model.Session{ID: record.ID, UserID: record.UserID, ExpiresAt: record.ExpiresAt, RevokedAt: record.RevokedAt, CreatedAt: record.CreatedAt}, nil
}

func (s *Store) RevokeSession(ctx context.Context, id string, revokedAt time.Time) error {
	if err := s.db.WithContext(ctx).Model(&sessionRecord{}).Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", revokedAt).Error; err != nil {
		return fmt.Errorf("revoke session: %w", storageError(err))
	}
	return nil
}

func (s *Store) RevokeOtherSessions(ctx context.Context, userID, exceptSessionID string, revokedAt time.Time) error {
	err := s.db.WithContext(ctx).Model(&sessionRecord{}).
		Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, exceptSessionID).
		Update("revoked_at", revokedAt).Error
	if err != nil {
		return fmt.Errorf("revoke other sessions: %w", storageError(err))
	}
	return nil
}

func (s *Store) Overview(ctx context.Context) (model.Overview, error) {
	var overview model.Overview
	for table, destination := range map[string]*int64{
		"students": new(int64), "classes": new(int64), "assignments": new(int64), "exams": new(int64),
	} {
		if err := s.db.WithContext(ctx).Table(table).Where("user_id = ?", userID(ctx)).Count(destination).Error; err != nil {
			return model.Overview{}, fmt.Errorf("count %s: %w", table, err)
		}
		switch table {
		case "students":
			overview.Students = int(*destination)
		case "classes":
			overview.Classes = int(*destination)
		case "assignments":
			overview.Assignments = int(*destination)
		case "exams":
			overview.Exams = int(*destination)
		}
	}
	var school schoolRecord
	if err := s.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID(ctx), true).Order("id").First(&school).Error; err != nil {
		return model.Overview{}, storageError(err)
	}
	overview.School = model.Reference{ID: school.ID, Name: school.Name}
	var year academicYearRecord
	if err := s.db.WithContext(ctx).Where("user_id = ? AND school_id = ? AND is_current = ?", userID(ctx), school.ID, true).First(&year).Error; err != nil {
		return model.Overview{}, storageError(err)
	}
	overview.AcademicYear = model.Reference{ID: year.ID, Name: year.Name}
	return overview, nil
}
