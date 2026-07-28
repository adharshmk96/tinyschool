package gormsqlite

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func examPreloads(db *gorm.DB) *gorm.DB {
	return db.Preload("Class").Preload("Students.Student")
}

func (s *Store) ListExams(ctx context.Context, options storage.ListOptions) ([]model.Exam, int64, error) {
	query := s.db.WithContext(ctx).Model(&examRecord{}).Where("exams.user_id = ?", userID(ctx))
	if options.Search != "" {
		p := contains(options.Search)
		query = query.Where("LOWER(exams.name) LIKE ? OR LOWER(exams.type) LIKE ? OR EXISTS (SELECT 1 FROM classes c WHERE c.id = exams.class_id AND LOWER(c.name) LIKE ?)", p, p, p)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowed := map[string]string{
		"name": "name", "type": "type", "examDate": "exam_date",
		"markedCount":  "(SELECT COUNT(*) FROM exam_students es WHERE es.exam_id = exams.id AND es.marked_at IS NOT NULL)",
		"averageScore": "(SELECT COALESCE(AVG(score), 0) FROM exam_students es WHERE es.exam_id = exams.id AND es.score IS NOT NULL)",
	}
	var records []examRecord
	err := examPreloads(paginate(order(query, options, allowed, "exam_date"), options)).Find(&records).Error
	items := make([]model.Exam, len(records))
	for i := range records {
		items[i] = examModel(records[i])
	}
	return items, total, storageError(err)
}

func (s *Store) Exam(ctx context.Context, id string) (model.Exam, error) {
	var record examRecord
	err := examPreloads(s.db.WithContext(ctx)).First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error
	return examModel(record), storageError(err)
}

func examRecordFrom(ownerID string, m model.Exam) examRecord {
	return examRecord{ID: m.ID, UserID: ownerID, SchoolID: m.SchoolID, AcademicYearID: m.AcademicYearID, ClassID: m.ClassID, Name: m.Name, Type: m.Type, ExamDate: m.ExamDate, TotalScore: m.TotalScore}
}

func replaceExamRoster(tx *gorm.DB, examID, classID string) error {
	if err := tx.Where("exam_id = ? AND score IS NULL AND marked_at IS NULL", examID).Delete(&examStudentRecord{}).Error; err != nil {
		return err
	}
	var studentIDs []string
	if err := tx.Model(&classStudentRecord{}).Where("class_id = ?", classID).Pluck("student_id", &studentIDs).Error; err != nil {
		return err
	}
	for _, id := range studentIDs {
		row := examStudentRecord{ExamID: examID, StudentID: id}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateExam(ctx context.Context, exam model.Exam) (model.Exam, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&[]examRecord{examRecordFrom(userID(ctx), exam)}).Error; err != nil {
			return err
		}
		return replaceExamRoster(tx, exam.ID, exam.ClassID)
	})
	return exam, storageError(err)
}

func (s *Store) UpdateExam(ctx context.Context, exam model.Exam) (model.Exam, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var old examRecord
		if err := tx.First(&old, "id = ? AND user_id = ?", exam.ID, userID(ctx)).Error; err != nil {
			return err
		}
		if err := tx.Model(&old).Updates(map[string]any{
			"school_id": exam.SchoolID, "academic_year_id": exam.AcademicYearID, "class_id": exam.ClassID,
			"name": exam.Name, "type": exam.Type, "exam_date": exam.ExamDate, "total_score": exam.TotalScore,
		}).Error; err != nil {
			return err
		}
		if old.ClassID != exam.ClassID {
			return replaceExamRoster(tx, exam.ID, exam.ClassID)
		}
		return nil
	})
	return exam, storageError(err)
}

func (s *Store) DeleteExam(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&examRecord{}, "id = ? AND user_id = ?", id, userID(ctx))
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Store) SetExamScore(ctx context.Context, examID, studentID string, score *float64, markedAt *time.Time) error {
	result := s.db.WithContext(ctx).Model(&examStudentRecord{}).
		Where("exam_id = ? AND student_id = ? AND EXISTS (SELECT 1 FROM exams WHERE exams.id = exam_students.exam_id AND exams.user_id = ?)", examID, studentID, userID(ctx)).
		Updates(map[string]any{"score": score, "marked_at": markedAt})
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}
