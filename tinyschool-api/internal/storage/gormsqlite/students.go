package gormsqlite

import (
	"context"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func (s *Store) ListStudents(ctx context.Context, options storage.ListOptions) ([]model.Student, int64, error) {
	// Students are shared across academic years; only the grade shown alongside
	// them follows the school's current year.
	query := s.db.WithContext(ctx).Model(&studentRecord{}).
		Joins("LEFT JOIN academic_years cay ON cay.school_id = students.school_id AND cay.user_id = students.user_id AND cay.is_current = 1").
		Joins("LEFT JOIN student_grades sg ON sg.student_id = students.id AND sg.academic_year_id = cay.id").
		Where("students.user_id = ?", userID(ctx))
	if options.Search != "" {
		p := contains(options.Search)
		query = query.Where("LOWER(first_name || ' ' || last_name) LIKE ? OR LOWER(students.email) LIKE ? OR LOWER(students.phone) LIKE ? OR LOWER(sg.grade) LIKE ? OR LOWER(guardian_name) LIKE ?", p, p, p, p, p)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowed := map[string]string{"name": "first_name || ' ' || last_name", "email": "students.email", "grade": "sg.grade", "averageScore": "(SELECT COALESCE(AVG(score * 100.0 / total_score), 0) FROM assignment_students ast JOIN assignments a ON a.id = ast.assignment_id WHERE ast.student_id = students.id AND ast.score IS NOT NULL)"}
	var records []studentRecord
	err := paginate(order(query, options, allowed, "first_name || ' ' || last_name"), options).
		Select("students.*").Preload("Grades.AcademicYear").Find(&records).Error
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
	if err := s.db.WithContext(ctx).Preload("Grades.AcademicYear").First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error; err != nil {
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
		Where("classes.user_id = ? AND class_students.student_id = ?", userID(ctx), id).
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
		if err := s.db.WithContext(ctx).First(&a, "id = ? AND user_id = ?", row.AssignmentID, userID(ctx)).Error; err != nil {
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
		if err := s.db.WithContext(ctx).First(&exam, "id = ? AND user_id = ?", row.ExamID, userID(ctx)).Error; err != nil {
			return model.Student{}, err
		}
		result.Exams = append(result.Exams, model.Result{ID: exam.ID, Name: exam.Name, Kind: "exam", DueAt: exam.ExamDate, Score: row.Score, TotalScore: exam.TotalScore, CompletedAt: row.MarkedAt})
	}
	return result, nil
}

func studentRecordFrom(ownerID string, m model.Student) studentRecord {
	return studentRecord{ID: m.ID, UserID: ownerID, SchoolID: m.SchoolID, FirstName: m.FirstName, LastName: m.LastName, Email: m.Email, Phone: m.Phone, GuardianName: m.GuardianName, GuardianEmail: m.GuardianEmail, GuardianPhone: m.GuardianPhone, ResidentAddress: m.ResidentAddress, PermanentAddress: m.PermanentAddress}
}

func replaceStudentGrades(tx *gorm.DB, studentID string, grades []model.StudentGrade) error {
	if err := tx.Where("student_id = ?", studentID).Delete(&studentGradeRecord{}).Error; err != nil {
		return err
	}
	for _, grade := range grades {
		row := studentGradeRecord{StudentID: studentID, AcademicYearID: grade.AcademicYearID, Grade: grade.Grade}
		if err := tx.Omit("AcademicYear").Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateStudent(ctx context.Context, student model.Student) (model.Student, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Grades").Create(&[]studentRecord{studentRecordFrom(userID(ctx), student)}).Error; err != nil {
			return err
		}
		return replaceStudentGrades(tx, student.ID, student.Grades)
	})
	return student, storageError(err)
}

func (s *Store) UpdateStudent(ctx context.Context, student model.Student) (model.Student, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&studentRecord{}).Where("id = ? AND user_id = ?", student.ID, userID(ctx)).Updates(map[string]any{
			"school_id": student.SchoolID, "first_name": student.FirstName, "last_name": student.LastName,
			"email": student.Email, "phone": student.Phone,
			"guardian_name": student.GuardianName, "guardian_email": student.GuardianEmail,
			"guardian_phone": student.GuardianPhone, "resident_address": student.ResidentAddress,
			"permanent_address": student.PermanentAddress,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return storage.ErrNotFound
		}
		return replaceStudentGrades(tx, student.ID, student.Grades)
	})
	if err != nil {
		return model.Student{}, storageError(err)
	}
	return student, nil
}

func (s *Store) DeleteStudent(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&studentRecord{}, "id = ? AND user_id = ?", id, userID(ctx))
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Store) CreateStudentLog(ctx context.Context, log model.StudentLog) (model.StudentLog, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&studentRecord{}).Where("id = ? AND user_id = ?", log.StudentID, userID(ctx)).Count(&count).Error; err != nil {
		return model.StudentLog{}, err
	}
	if count == 0 {
		return model.StudentLog{}, storage.ErrNotFound
	}
	record := studentLogRecord{ID: log.ID, StudentID: log.StudentID, Kind: log.Kind, Type: log.Type, Note: log.Note, CreatedAt: log.CreatedAt}
	return log, storageError(s.db.WithContext(ctx).Create(&record).Error)
}

func (s *Store) DeleteStudentLog(ctx context.Context, studentID, logID, kind string) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND student_id = ? AND kind = ? AND EXISTS (SELECT 1 FROM students WHERE students.id = student_logs.student_id AND students.user_id = ?)", logID, studentID, kind, userID(ctx)).
		Delete(&studentLogRecord{})
	if result.Error != nil {
		return storageError(result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrNotFound
	}
	return nil
}
