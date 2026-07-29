package gormsqlite

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func assignmentPreloads(db *gorm.DB) *gorm.DB {
	return db.Preload("Class").Preload("Students.Student.Classrooms.AcademicYear")
}

func (s *Store) ListAssignments(ctx context.Context, options storage.ListOptions) ([]model.Assignment, int64, error) {
	query := s.db.WithContext(ctx).Model(&assignmentRecord{}).Where("assignments.user_id = ?", userID(ctx))
	if options.AcademicYearID != "" {
		query = query.Where("assignments.academic_year_id = ?", options.AcademicYearID)
	}
	if options.Classroom != "" {
		classroom := options.Classroom
		yearID := options.AcademicYearID
		query = query.Where(`
			EXISTS (
				SELECT 1 FROM class_classrooms cc
				WHERE cc.class_id = assignments.class_id AND LOWER(cc.classroom) = LOWER(?)
			)
			OR EXISTS (
				SELECT 1 FROM assignment_students ast
				JOIN student_classrooms sc ON sc.student_id = ast.student_id
				WHERE ast.assignment_id = assignments.id
				  AND LOWER(sc.classroom) = LOWER(?)
				  AND (? = '' OR sc.academic_year_id = ?)
			)`, classroom, classroom, yearID, yearID)
	}
	if options.Search != "" {
		p := contains(options.Search)
		query = query.Where("LOWER(assignments.name) LIKE ? OR LOWER(assignments.type) LIKE ? OR EXISTS (SELECT 1 FROM classes c WHERE c.id = assignments.class_id AND LOWER(c.name) LIKE ?)", p, p, p)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowed := map[string]string{
		"name": "name", "type": "type", "dueDate": "due_date",
		"completion": "(SELECT CASE WHEN COUNT(*) = 0 THEN 0 ELSE 100 * SUM(CASE WHEN completed_at IS NOT NULL THEN 1 ELSE 0 END) / COUNT(*) END FROM assignment_students ast WHERE ast.assignment_id = assignments.id)",
	}
	var records []assignmentRecord
	err := assignmentPreloads(paginate(order(query, options, allowed, "due_date"), options)).Find(&records).Error
	items := make([]model.Assignment, len(records))
	for i := range records {
		items[i] = assignmentModel(records[i])
	}
	return items, total, storageError(err)
}

func (s *Store) Assignment(ctx context.Context, id string) (model.Assignment, error) {
	var record assignmentRecord
	err := assignmentPreloads(s.db.WithContext(ctx)).First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error
	return assignmentModel(record), storageError(err)
}

func assignmentRecordFrom(ownerID string, m model.Assignment) assignmentRecord {
	return assignmentRecord{ID: m.ID, UserID: ownerID, SchoolID: m.SchoolID, AcademicYearID: m.AcademicYearID, Name: m.Name, Type: m.Type, DueDate: m.DueDate, TotalScore: m.TotalScore, ClassID: m.ClassID}
}

func assignmentRoster(tx *gorm.DB, assignment model.Assignment, studentIDs []string) error {
	if assignment.ClassID != nil {
		studentIDs = nil
		if err := tx.Model(&classStudentRecord{}).Where("class_id = ?", *assignment.ClassID).Pluck("student_id", &studentIDs).Error; err != nil {
			return err
		}
	}
	for _, id := range studentIDs {
		row := assignmentStudentRecord{AssignmentID: assignment.ID, StudentID: id}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateAssignment(ctx context.Context, assignment model.Assignment, studentIDs []string) (model.Assignment, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&[]assignmentRecord{assignmentRecordFrom(userID(ctx), assignment)}).Error; err != nil {
			return err
		}
		return assignmentRoster(tx, assignment, studentIDs)
	})
	return assignment, storageError(err)
}

func (s *Store) UpdateAssignment(ctx context.Context, assignment model.Assignment, studentIDs *[]string) (model.Assignment, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&assignmentRecord{}).Where("id = ? AND user_id = ?", assignment.ID, userID(ctx)).Updates(map[string]any{
			"school_id": assignment.SchoolID, "academic_year_id": assignment.AcademicYearID,
			"name": assignment.Name, "type": assignment.Type, "due_date": assignment.DueDate,
			"total_score": assignment.TotalScore, "class_id": assignment.ClassID,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if studentIDs == nil && assignment.ClassID == nil {
			return nil
		}
		if err := tx.Where("assignment_id = ? AND score IS NULL AND completed_at IS NULL", assignment.ID).Delete(&assignmentStudentRecord{}).Error; err != nil {
			return err
		}
		ids := []string(nil)
		if studentIDs != nil {
			ids = *studentIDs
		}
		return assignmentRoster(tx, assignment, ids)
	})
	return assignment, storageError(err)
}

func (s *Store) DeleteAssignment(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&assignmentRecord{}, "id = ? AND user_id = ?", id, userID(ctx))
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Store) SetAssignmentScore(ctx context.Context, assignmentID, studentID string, score *float64, completedAt *time.Time) error {
	result := s.db.WithContext(ctx).Model(&assignmentStudentRecord{}).
		Where("assignment_id = ? AND student_id = ? AND EXISTS (SELECT 1 FROM assignments WHERE assignments.id = assignment_students.assignment_id AND assignments.user_id = ?)", assignmentID, studentID, userID(ctx)).
		Updates(map[string]any{"score": score, "completed_at": completedAt})
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}
