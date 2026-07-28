package gormsqlite

import (
	"context"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func (s *Store) ListSchools(ctx context.Context, options storage.ListOptions) ([]model.School, int64, error) {
	query := s.db.WithContext(ctx).Model(&schoolRecord{}).Where("schools.user_id = ?", userID(ctx))
	if options.Search != "" {
		search := contains(options.Search)
		query = query.Where("LOWER(name) LIKE ? OR EXISTS (SELECT 1 FROM school_grades g WHERE g.school_id = schools.id AND LOWER(g.grade) LIKE ?)", search, search)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []schoolRecord
	err := paginate(order(query, options, map[string]string{"name": "name"}, "name"), options).
		Preload("Grades", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).Find(&records).Error
	items := make([]model.School, len(records))
	for i := range records {
		items[i] = schoolModel(records[i])
	}
	return items, total, storageError(err)
}

func (s *Store) School(ctx context.Context, id string) (model.School, error) {
	var record schoolRecord
	err := s.db.WithContext(ctx).Preload("Grades", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error
	return schoolModel(record), storageError(err)
}

func createSchool(tx *gorm.DB, ownerID string, school model.School) error {
	record := schoolRecord{ID: school.ID, UserID: ownerID, Name: school.Name, IsActive: school.IsActive}
	for i, grade := range school.Grades {
		record.Grades = append(record.Grades, schoolGradeRecord{SchoolID: school.ID, Grade: grade, Position: i})
	}
	return tx.Create(&record).Error
}

func (s *Store) CreateSchool(ctx context.Context, school model.School) (model.School, error) {
	err := createSchool(s.db.WithContext(ctx), userID(ctx), school)
	return school, storageError(err)
}

func (s *Store) UpdateSchool(ctx context.Context, school model.School) (model.School, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&schoolRecord{}).Where("id = ? AND user_id = ?", school.ID, userID(ctx)).Updates(map[string]any{"name": school.Name, "is_active": school.IsActive})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("school_id = ?", school.ID).Delete(&schoolGradeRecord{}).Error; err != nil {
			return err
		}
		for i, grade := range school.Grades {
			if err := tx.Create(&schoolGradeRecord{SchoolID: school.ID, Grade: grade, Position: i}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return school, storageError(err)
}

func (s *Store) DeleteSchool(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&schoolRecord{}, "id = ? AND user_id = ?", id, userID(ctx))
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}
