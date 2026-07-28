package gormsqlite

import (
	"context"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func (s *Store) ListStudents(ctx context.Context, options storage.ListOptions) ([]model.Student, int64, error) {
	query := s.db.WithContext(ctx).Model(&studentRecord{})
	if options.Search != "" {
		p := contains(options.Search)
		query = query.Where("LOWER(first_name || ' ' || last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(grade) LIKE ? OR LOWER(guardian_name) LIKE ?", p, p, p, p, p)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowed := map[string]string{"name": "first_name || ' ' || last_name", "email": "email", "grade": "grade", "averageScore": "(SELECT COALESCE(AVG(score * 100.0 / total_score), 0) FROM assignment_students ast JOIN assignments a ON a.id = ast.assignment_id WHERE ast.student_id = students.id AND ast.score IS NOT NULL)"}
	var records []studentRecord
	err := paginate(order(query, options, allowed, "first_name || ' ' || last_name"), options).Find(&records).Error
	items := make([]model.Student, len(records))
	for i := range records {
		items[i] = studentModel(records[i])
	}
	return items, total, storageError(err)
}

func preloadStudent(db *gorm.DB) *gorm.DB {
	return db.Preload("Logs", func(db *gorm.DB) *gorm.DB { return db.Order("created_at DESC") })
}

func (s *Store) Student(ctx context.Context, id string) (model.Student, error) {
	var record studentRecord
	if err := s.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return model.Student{}, storageError(err)
	}
	result := studentModel(record)
	var logs []studentLogRecord
	if err := s.db.WithContext(ctx).Where("student_id = ?", id).Order("created_at DESC").Find(&logs).Error; err != nil {
		return model.Student{}, err
	}
	for _, log := range logs {
		result.Logs = append(result.Logs, model.StudentLog{ID: log.ID, StudentID: log.StudentID, Kind: log.Kind, Type: log.Type, Note: log.Note, CreatedAt: log.CreatedAt})
	}
	var classes []classRecord
	if err := s.db.WithContext(ctx).
		Joins("JOIN class_students ON class_students.class_id = classes.id").
		Where("class_students.student_id = ?", id).
		Order("classes.name").
		Find(&classes).Error; err != nil {
		return model.Student{}, err
	}
	for _, class := range classes {
		result.Classes = append(result.Classes, model.Reference{ID: class.ID, Name: class.Name})
	}
	var assignments []assignmentStudentRecord
	if err := s.db.WithContext(ctx).Preload("Student").Where("student_id = ?", id).Find(&assignments).Error; err != nil {
		return model.Student{}, err
	}
	for _, row := range assignments {
		var a assignmentRecord
		if err := s.db.WithContext(ctx).First(&a, "id = ?", row.AssignmentID).Error; err != nil {
			return model.Student{}, err
		}
		result.Assignments = append(result.Assignments, model.Result{ID: a.ID, Name: a.Name, Kind: "assignment", DueAt: a.DueDate, Score: row.Score, TotalScore: a.TotalScore, CompletedAt: row.CompletedAt})
	}
	var exams []examStudentRecord
	if err := s.db.WithContext(ctx).Where("student_id = ?", id).Find(&exams).Error; err != nil {
		return model.Student{}, err
	}
	for _, row := range exams {
		var exam examRecord
		if err := s.db.WithContext(ctx).First(&exam, "id = ?", row.ExamID).Error; err != nil {
			return model.Student{}, err
		}
		result.Exams = append(result.Exams, model.Result{ID: exam.ID, Name: exam.Name, Kind: "exam", DueAt: exam.ExamDate, Score: row.Score, TotalScore: exam.TotalScore, CompletedAt: row.MarkedAt})
	}
	return result, nil
}

func studentRecordFrom(m model.Student) studentRecord {
	return studentRecord{ID: m.ID, SchoolID: m.SchoolID, FirstName: m.FirstName, LastName: m.LastName, Email: m.Email, Phone: m.Phone, Grade: m.Grade, GuardianName: m.GuardianName, GuardianEmail: m.GuardianEmail, GuardianPhone: m.GuardianPhone, ResidentAddress: m.ResidentAddress, PermanentAddress: m.PermanentAddress}
}

func (s *Store) CreateStudent(ctx context.Context, student model.Student) (model.Student, error) {
	err := s.db.WithContext(ctx).Create(&[]studentRecord{studentRecordFrom(student)}).Error
	return student, storageError(err)
}

func (s *Store) UpdateStudent(ctx context.Context, student model.Student) (model.Student, error) {
	result := s.db.WithContext(ctx).Model(&studentRecord{}).Where("id = ?", student.ID).Updates(map[string]any{
		"school_id": student.SchoolID, "first_name": student.FirstName, "last_name": student.LastName,
		"email": student.Email, "phone": student.Phone, "grade": student.Grade,
		"guardian_name": student.GuardianName, "guardian_email": student.GuardianEmail,
		"guardian_phone": student.GuardianPhone, "resident_address": student.ResidentAddress,
		"permanent_address": student.PermanentAddress,
	})
	if result.Error != nil {
		return model.Student{}, storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return model.Student{}, storage.ErrNotFound
	}
	return student, nil
}

func (s *Store) DeleteStudent(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&studentRecord{}, "id = ?", id)
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Store) CreateStudentLog(ctx context.Context, log model.StudentLog) (model.StudentLog, error) {
	record := studentLogRecord{ID: log.ID, StudentID: log.StudentID, Kind: log.Kind, Type: log.Type, Note: log.Note, CreatedAt: log.CreatedAt}
	return log, storageError(s.db.WithContext(ctx).Create(&record).Error)
}

func (s *Store) DeleteStudentLog(ctx context.Context, studentID, logID, kind string) error {
	result := s.db.WithContext(ctx).Where("id = ? AND student_id = ? AND kind = ?", logID, studentID, kind).Delete(&studentLogRecord{})
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}
