package gormsqlite

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func (s *Store) ListSchools(ctx context.Context, options storage.ListOptions) ([]model.School, int64, error) {
	query := s.db.WithContext(ctx).Model(&schoolRecord{}).Where("schools.user_id = ?", userID(ctx))
	if options.Search != "" {
		search := contains(options.Search)
		query = query.Where("LOWER(name) LIKE ? OR EXISTS (SELECT 1 FROM school_classrooms c WHERE c.school_id = schools.id AND LOWER(c.classroom) LIKE ?)", search, search)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []schoolRecord
	err := paginate(order(query, options, map[string]string{"name": "name"}, "name"), options).
		Preload("Classrooms", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).Find(&records).Error
	items := make([]model.School, len(records))
	for i := range records {
		items[i] = schoolModel(records[i])
	}
	return items, total, storageError(err)
}

func (s *Store) School(ctx context.Context, id string) (model.School, error) {
	var record schoolRecord
	err := s.db.WithContext(ctx).Preload("Classrooms", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error
	return schoolModel(record), storageError(err)
}

// SchoolClassroomsInUse returns distinct classroom strings linked to classes or students for the school.
func (s *Store) SchoolClassroomsInUse(ctx context.Context, schoolID string) ([]string, error) {
	ownerID := userID(ctx)
	var classClassrooms []string
	if err := s.db.WithContext(ctx).Model(&classClassroomRecord{}).
		Joins("JOIN classes ON classes.id = class_classrooms.class_id").
		Where("classes.school_id = ? AND classes.user_id = ?", schoolID, ownerID).
		Distinct().Pluck("class_classrooms.classroom", &classClassrooms).Error; err != nil {
		return nil, storageError(err)
	}
	var studentClassrooms []string
	if err := s.db.WithContext(ctx).Model(&studentClassroomRecord{}).
		Joins("JOIN students ON students.id = student_classrooms.student_id").
		Where("students.school_id = ? AND students.user_id = ?", schoolID, ownerID).
		Distinct().Pluck("student_classrooms.classroom", &studentClassrooms).Error; err != nil {
		return nil, storageError(err)
	}
	seen := make(map[string]struct{}, len(classClassrooms)+len(studentClassrooms))
	result := make([]string, 0, len(classClassrooms)+len(studentClassrooms))
	for _, classroom := range append(classClassrooms, studentClassrooms...) {
		key := strings.ToLower(strings.TrimSpace(classroom))
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, classroom)
	}
	return result, nil
}

func createSchool(tx *gorm.DB, ownerID string, school model.School) error {
	record := schoolRecord{ID: school.ID, UserID: ownerID, Name: school.Name, IsActive: school.IsActive}
	for i, classroom := range school.Classrooms {
		record.Classrooms = append(record.Classrooms, schoolClassroomRecord{SchoolID: school.ID, Classroom: classroom, Position: i})
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
		if err := tx.Where("school_id = ?", school.ID).Delete(&schoolClassroomRecord{}).Error; err != nil {
			return err
		}
		for i, classroom := range school.Classrooms {
			if err := tx.Create(&schoolClassroomRecord{SchoolID: school.ID, Classroom: classroom, Position: i}).Error; err != nil {
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
