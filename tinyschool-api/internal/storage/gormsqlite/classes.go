package gormsqlite

import (
	"context"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func (s *Store) ListClasses(ctx context.Context, options storage.ListOptions) ([]model.Class, int64, error) {
	query := s.db.WithContext(ctx).Model(&classRecord{}).Where("classes.user_id = ?", userID(ctx))
	if options.Search != "" {
		p := contains(options.Search)
		query = query.Where("LOWER(name) LIKE ? OR LOWER(subject) LIKE ? OR LOWER(grade) LIKE ? OR LOWER(description) LIKE ?", p, p, p, p)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowed := map[string]string{"name": "name", "subject": "subject", "grade": "grade", "studentCount": "(SELECT COUNT(*) FROM class_students cs WHERE cs.class_id = classes.id)"}
	var records []classRecord
	err := paginate(order(query, options, allowed, "name"), options).Preload("Students").Find(&records).Error
	items := make([]model.Class, len(records))
	for i := range records {
		items[i] = classModel(records[i])
	}
	return items, total, storageError(err)
}

func (s *Store) Class(ctx context.Context, id string) (model.Class, error) {
	var record classRecord
	if err := s.db.WithContext(ctx).Preload("Students").First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error; err != nil {
		return model.Class{}, storageError(err)
	}
	result := classModel(record)
	var assignments []assignmentRecord
	if err := s.db.WithContext(ctx).Where("user_id = ? AND class_id = ?", userID(ctx), id).Order("name").Find(&assignments).Error; err != nil {
		return model.Class{}, err
	}
	for _, a := range assignments {
		result.Assignments = append(result.Assignments, model.Reference{ID: a.ID, Name: a.Name})
	}
	var exams []examRecord
	if err := s.db.WithContext(ctx).Where("user_id = ? AND class_id = ?", userID(ctx), id).Order("name").Find(&exams).Error; err != nil {
		return model.Class{}, err
	}
	for _, exam := range exams {
		result.Exams = append(result.Exams, model.Reference{ID: exam.ID, Name: exam.Name})
	}
	return result, nil
}

func classRecordFrom(ownerID string, m model.Class) classRecord {
	return classRecord{ID: m.ID, UserID: ownerID, SchoolID: m.SchoolID, AcademicYearID: m.AcademicYearID, Name: m.Name, Subject: m.Subject, Grade: m.Grade, Description: m.Description}
}

func replaceClassStudents(tx *gorm.DB, classID string, studentIDs []string) error {
	if err := tx.Where("class_id = ?", classID).Delete(&classStudentRecord{}).Error; err != nil {
		return err
	}
	for _, studentID := range studentIDs {
		if err := tx.Create(&classStudentRecord{ClassID: classID, StudentID: studentID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateClass(ctx context.Context, class model.Class, studentIDs []string) (model.Class, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&[]classRecord{classRecordFrom(userID(ctx), class)}).Error; err != nil {
			return err
		}
		return replaceClassStudents(tx, class.ID, studentIDs)
	})
	return class, storageError(err)
}

func (s *Store) UpdateClass(ctx context.Context, class model.Class, studentIDs *[]string) (model.Class, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&classRecord{}).Where("id = ? AND user_id = ?", class.ID, userID(ctx)).Updates(map[string]any{
			"school_id": class.SchoolID, "academic_year_id": class.AcademicYearID, "name": class.Name,
			"subject": class.Subject, "grade": class.Grade, "description": class.Description,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if studentIDs != nil {
			return replaceClassStudents(tx, class.ID, *studentIDs)
		}
		return nil
	})
	return class, storageError(err)
}

func (s *Store) DeleteClass(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&classRecord{}, "id = ? AND user_id = ?", id, userID(ctx))
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}
