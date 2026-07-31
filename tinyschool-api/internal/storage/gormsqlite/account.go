package gormsqlite

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
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
	role := user.Role
	if role == "" {
		role = model.RoleUser
	}
	user.Role = role
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := &userRecord{
			ID: user.ID, Name: user.Name, Email: user.Email,
			PasswordHash: user.PasswordHash, Role: role, CreatedAt: user.CreatedAt,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		user.CreatedAt = record.CreatedAt
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

func (s *Store) CountAdmins(ctx context.Context) (int64, error) {
	var total int64
	if err := s.db.WithContext(ctx).Model(&userRecord{}).Where("role = ?", model.RoleAdmin).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return total, nil
}

var userSortColumns = map[string]string{
	"name": "name", "email": "email", "createdAt": "created_at", "role": "role",
}

// ListUsers is the only query that deliberately ignores tenancy: the back
// office browses every account in the database.
func (s *Store) ListUsers(ctx context.Context, options storage.ListOptions) ([]model.User, int64, error) {
	query := s.db.WithContext(ctx).Model(&userRecord{})
	if options.Search != "" {
		pattern := contains(options.Search)
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	var records []userRecord
	err := paginate(order(query, options, userSortColumns, "name"), options).Find(&records).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	users := make([]model.User, len(records))
	for i := range records {
		users[i] = userModel(records[i])
	}
	return users, total, nil
}

func (s *Store) SetUserBlocked(ctx context.Context, id string, blockedAt *time.Time) (model.User, error) {
	result := s.db.WithContext(ctx).Model(&userRecord{}).Where("id = ?", id).
		Update("blocked_at", blockedAt)
	if result.Error != nil {
		return model.User{}, storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return model.User{}, storage.ErrNotFound
	}
	return s.UserByID(ctx, id)
}

func (s *Store) RevokeUserSessions(ctx context.Context, userID string, revokedAt time.Time) error {
	err := s.db.WithContext(ctx).Model(&sessionRecord{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", storageError(err))
	}
	return nil
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

func (s *Store) CreatePasswordResetToken(ctx context.Context, token model.PasswordResetToken) (model.PasswordResetToken, error) {
	record := passwordResetRecord{
		ID: token.ID, UserID: token.UserID, TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt, UsedAt: token.UsedAt, CreatedAt: token.CreatedAt,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return model.PasswordResetToken{}, storageError(err)
	}
	return token, nil
}

func (s *Store) PasswordResetTokenByHash(ctx context.Context, hash string) (model.PasswordResetToken, error) {
	var record passwordResetRecord
	if err := s.db.WithContext(ctx).First(&record, "token_hash = ?", hash).Error; err != nil {
		return model.PasswordResetToken{}, storageError(err)
	}
	return model.PasswordResetToken{
		ID: record.ID, UserID: record.UserID, TokenHash: record.TokenHash,
		ExpiresAt: record.ExpiresAt, UsedAt: record.UsedAt, CreatedAt: record.CreatedAt,
	}, nil
}

func (s *Store) UsePasswordResetToken(ctx context.Context, id string, usedAt time.Time) error {
	result := s.db.WithContext(ctx).Model(&passwordResetRecord{}).
		Where("id = ? AND used_at IS NULL", id).Update("used_at", usedAt)
	if result.Error != nil {
		return fmt.Errorf("use password reset token: %w", storageError(result.Error))
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Store) InvalidateUserPasswordResetTokens(ctx context.Context, userID string, usedAt time.Time) error {
	err := s.db.WithContext(ctx).Model(&passwordResetRecord{}).
		Where("user_id = ? AND used_at IS NULL", userID).Update("used_at", usedAt).Error
	if err != nil {
		return fmt.Errorf("invalidate password reset tokens: %w", storageError(err))
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

func (s *Store) Upcoming(ctx context.Context, academicYearID, fromDate string, limit int) ([]model.UpcomingItem, error) {
	if limit < 1 {
		return []model.UpcomingItem{}, nil
	}

	var assignments []assignmentRecord
	err := assignmentPreloads(s.db.WithContext(ctx)).
		Where("assignments.user_id = ? AND assignments.academic_year_id = ? AND assignments.due_date >= ?", userID(ctx), academicYearID, fromDate).
		Order("assignments.due_date ASC, assignments.name ASC").
		Limit(limit).
		Find(&assignments).Error
	if err != nil {
		return nil, storageError(err)
	}

	var exams []examRecord
	err = examPreloads(s.db.WithContext(ctx)).
		Where("exams.user_id = ? AND exams.academic_year_id = ? AND exams.exam_date >= ?", userID(ctx), academicYearID, fromDate).
		Order("exams.exam_date ASC, exams.name ASC").
		Limit(limit).
		Find(&exams).Error
	if err != nil {
		return nil, storageError(err)
	}

	items := make([]model.UpcomingItem, 0, len(assignments)+len(exams))
	for _, record := range assignments {
		item := model.UpcomingItem{
			ID: record.ID, Kind: "assignment", Name: record.Name, Date: record.DueDate,
			StudentCount: len(record.Students),
		}
		if record.Class != nil {
			item.Class = &model.Reference{ID: record.Class.ID, Name: record.Class.Name}
		}
		items = append(items, item)
	}
	for _, record := range exams {
		items = append(items, model.UpcomingItem{
			ID: record.ID, Kind: "exam", Name: record.Name, Date: record.ExamDate,
			Class:        &model.Reference{ID: record.Class.ID, Name: record.Class.Name},
			StudentCount: len(record.Students),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Date == items[j].Date {
			return items[i].Name < items[j].Name
		}
		return items[i].Date < items[j].Date
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}
