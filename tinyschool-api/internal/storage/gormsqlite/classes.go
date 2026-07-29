package gormsqlite

import (
	"context"
	"sort"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

func (s *Store) ListClasses(ctx context.Context, options storage.ListOptions) ([]model.Class, int64, error) {
	query := s.db.WithContext(ctx).Model(&classRecord{}).Where("classes.user_id = ?", userID(ctx))
	if options.AcademicYearID != "" {
		query = query.Where("classes.academic_year_id = ?", options.AcademicYearID)
	}
	if options.Classroom != "" {
		query = query.Where(`EXISTS (
			SELECT 1 FROM class_classrooms cc
			WHERE cc.class_id = classes.id AND LOWER(cc.classroom) = LOWER(?)
		)`, options.Classroom)
	}
	if options.Search != "" {
		p := contains(options.Search)
		query = query.Where(`LOWER(name) LIKE ? OR LOWER(subject) LIKE ? OR LOWER(description) LIKE ?
			OR EXISTS (SELECT 1 FROM class_classrooms cc WHERE cc.class_id = classes.id AND LOWER(cc.classroom) LIKE ?)`, p, p, p, p)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowed := map[string]string{
		"name": "name", "subject": "subject",
		"classroom": "(SELECT MIN(cc.classroom) FROM class_classrooms cc WHERE cc.class_id = classes.id)",
		"studentCount": "(SELECT COUNT(*) FROM class_students cs WHERE cs.class_id = classes.id)",
	}
	var records []classRecord
	err := paginate(order(query, options, allowed, "name"), options).
		Preload("Classrooms").
		Preload("Students.Classrooms.AcademicYear").
		Find(&records).Error
	items := make([]model.Class, len(records))
	for i := range records {
		items[i] = classModel(records[i])
	}
	return items, total, storageError(err)
}

func (s *Store) Class(ctx context.Context, id string) (model.Class, error) {
	var record classRecord
	if err := s.db.WithContext(ctx).
		Preload("Classrooms").
		Preload("Students.Classrooms.AcademicYear").
		First(&record, "id = ? AND user_id = ?", id, userID(ctx)).Error; err != nil {
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
	var assignmentScores []assignmentStudentRecord
	if err := s.db.WithContext(ctx).
		Joins("JOIN assignments ON assignments.id = assignment_students.assignment_id").
		Where("assignments.user_id = ? AND assignments.class_id = ?", userID(ctx), id).
		Find(&assignmentScores).Error; err != nil {
		return model.Class{}, err
	}
	assignmentByID := make(map[string]assignmentRecord, len(assignments))
	for _, assignment := range assignments {
		assignmentByID[assignment.ID] = assignment
	}
	studentIndex := make(map[string]int, len(result.Students))
	for index, student := range result.Students {
		studentIndex[student.ID] = index
	}
	for _, row := range assignmentScores {
		assignment, ok := assignmentByID[row.AssignmentID]
		index, enrolled := studentIndex[row.StudentID]
		if !ok || !enrolled {
			continue
		}
		result.Students[index].Assignments = append(result.Students[index].Assignments, model.Result{
			ID: assignment.ID, Name: assignment.Name, Kind: "assignment", DueAt: assignment.DueDate,
			Score: row.Score, TotalScore: assignment.TotalScore, CompletedAt: row.CompletedAt,
		})
	}
	var exams []examRecord
	if err := s.db.WithContext(ctx).Where("user_id = ? AND class_id = ?", userID(ctx), id).Order("name").Find(&exams).Error; err != nil {
		return model.Class{}, err
	}
	for _, exam := range exams {
		result.Exams = append(result.Exams, model.Reference{ID: exam.ID, Name: exam.Name})
	}
	var examScores []examStudentRecord
	if err := s.db.WithContext(ctx).
		Joins("JOIN exams ON exams.id = exam_students.exam_id").
		Where("exams.user_id = ? AND exams.class_id = ?", userID(ctx), id).
		Find(&examScores).Error; err != nil {
		return model.Class{}, err
	}
	examByID := make(map[string]examRecord, len(exams))
	for _, exam := range exams {
		examByID[exam.ID] = exam
	}
	for _, row := range examScores {
		exam, ok := examByID[row.ExamID]
		index, enrolled := studentIndex[row.StudentID]
		if !ok || !enrolled {
			continue
		}
		result.Students[index].Exams = append(result.Students[index].Exams, model.Result{
			ID: exam.ID, Name: exam.Name, Kind: "exam", DueAt: exam.ExamDate,
			Score: row.Score, TotalScore: exam.TotalScore, CompletedAt: row.MarkedAt,
		})
	}
	for index := range result.Students {
		sort.Slice(result.Students[index].Exams, func(i, j int) bool {
			return result.Students[index].Exams[i].DueAt < result.Students[index].Exams[j].DueAt
		})
	}
	return result, nil
}

func classRecordFrom(ownerID string, m model.Class) classRecord {
	return classRecord{
		ID: m.ID, UserID: ownerID, SchoolID: m.SchoolID, AcademicYearID: m.AcademicYearID,
		Name: m.Name, Subject: m.Subject, Description: m.Description,
	}
}

func replaceClassClassrooms(tx *gorm.DB, classID string, classrooms []string) error {
	if err := tx.Where("class_id = ?", classID).Delete(&classClassroomRecord{}).Error; err != nil {
		return err
	}
	for _, classroom := range classrooms {
		if err := tx.Create(&classClassroomRecord{ClassID: classID, Classroom: classroom}).Error; err != nil {
			return err
		}
	}
	return nil
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
		if err := tx.Omit("Classrooms", "Students").Create(&[]classRecord{classRecordFrom(userID(ctx), class)}).Error; err != nil {
			return err
		}
		if err := replaceClassClassrooms(tx, class.ID, class.Classrooms); err != nil {
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
			"subject": class.Subject, "description": class.Description,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := replaceClassClassrooms(tx, class.ID, class.Classrooms); err != nil {
			return err
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
