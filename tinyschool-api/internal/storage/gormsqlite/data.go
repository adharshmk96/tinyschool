package gormsqlite

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"tinyschool-api/internal/model"
)

// ExportUserData reads every row owned by the caller in one pass. Unlike the
// list queries it never paginates and never joins for display purposes: the
// result is meant to round-trip back through ReplaceUserData unchanged.
func (s *Store) ExportUserData(ctx context.Context) (model.Dataset, error) {
	owner := userID(ctx)
	if owner == "" {
		return model.Dataset{}, fmt.Errorf("export user data: missing user identity")
	}
	db := s.db.WithContext(ctx)
	var dataset model.Dataset

	var schools []schoolRecord
	if err := db.Preload("Classrooms", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Where("user_id = ?", owner).Order("name").Find(&schools).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export schools: %w", err)
	}
	for _, school := range schools {
		row := model.SchoolRow{ID: school.ID, Name: school.Name, IsActive: school.IsActive}
		for _, classroom := range school.Classrooms {
			row.Classrooms = append(row.Classrooms, classroom.Classroom)
		}
		dataset.Schools = append(dataset.Schools, row)
	}

	var years []academicYearRecord
	if err := db.Preload("Segments", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Where("user_id = ?", owner).Order("start_date").Find(&years).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export academic years: %w", err)
	}
	for _, year := range years {
		dataset.AcademicYears = append(dataset.AcademicYears, model.AcademicYearRow{
			ID: year.ID, SchoolID: year.SchoolID, Name: year.Name,
			StartDate: year.StartDate, EndDate: year.EndDate, IsCurrent: year.IsCurrent,
		})
		for position, segment := range year.Segments {
			dataset.AcademicSegments = append(dataset.AcademicSegments, model.AcademicSegmentRow{
				ID: segment.ID, AcademicYearID: year.ID, Name: segment.Name, Type: segment.Type,
				StartDate: segment.StartDate, EndDate: segment.EndDate, Position: position,
			})
		}
	}

	var students []studentRecord
	if err := db.Preload("Classrooms").Where("user_id = ?", owner).
		Order("first_name, last_name").Find(&students).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export students: %w", err)
	}
	for _, student := range students {
		dataset.Students = append(dataset.Students, model.StudentRow{
			ID: student.ID, SchoolID: student.SchoolID, FirstName: student.FirstName, LastName: student.LastName,
			Email: student.Email, Phone: student.Phone, GuardianName: student.GuardianName,
			GuardianEmail: student.GuardianEmail, GuardianPhone: student.GuardianPhone,
			ResidentAddress: student.ResidentAddress, PermanentAddress: student.PermanentAddress,
		})
		for _, classroom := range student.Classrooms {
			dataset.StudentClassrooms = append(dataset.StudentClassrooms, model.StudentClassroomRow{
				StudentID: student.ID, AcademicYearID: classroom.AcademicYearID, Classroom: classroom.Classroom,
			})
		}
	}

	var logs []studentLogRecord
	if err := db.Joins("JOIN students ON students.id = student_logs.student_id").
		Where("students.user_id = ?", owner).Order("student_logs.created_at").Find(&logs).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export student logs: %w", err)
	}
	for _, log := range logs {
		dataset.StudentLogs = append(dataset.StudentLogs, model.StudentLogRow{
			ID: log.ID, StudentID: log.StudentID, Kind: log.Kind, Type: log.Type,
			Note: log.Note, CreatedAt: log.CreatedAt,
		})
	}

	var classes []classRecord
	if err := db.Preload("Classrooms").Where("user_id = ?", owner).Order("name").Find(&classes).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export classes: %w", err)
	}
	for _, class := range classes {
		row := model.ClassRow{
			ID: class.ID, SchoolID: class.SchoolID, AcademicYearID: class.AcademicYearID,
			Name: class.Name, Subject: class.Subject, Description: class.Description,
		}
		for _, classroom := range class.Classrooms {
			row.Classrooms = append(row.Classrooms, classroom.Classroom)
		}
		dataset.Classes = append(dataset.Classes, row)
	}

	var enrolments []classStudentRecord
	if err := db.Model(&classStudentRecord{}).
		Joins("JOIN classes ON classes.id = class_students.class_id").
		Where("classes.user_id = ?", owner).
		Order("class_students.class_id, class_students.student_id").Find(&enrolments).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export class students: %w", err)
	}
	for _, row := range enrolments {
		dataset.ClassStudents = append(dataset.ClassStudents, model.ClassStudentRow{ClassID: row.ClassID, StudentID: row.StudentID})
	}

	var assignments []assignmentRecord
	if err := db.Where("user_id = ?", owner).Order("due_date, name").Find(&assignments).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export assignments: %w", err)
	}
	for _, assignment := range assignments {
		row := model.AssignmentRow{
			ID: assignment.ID, SchoolID: assignment.SchoolID, AcademicYearID: assignment.AcademicYearID,
			Name: assignment.Name, Type: assignment.Type, DueDate: assignment.DueDate, TotalScore: assignment.TotalScore,
		}
		if assignment.ClassID != nil {
			row.ClassID = *assignment.ClassID
		}
		dataset.Assignments = append(dataset.Assignments, row)
	}

	var assignmentScores []assignmentStudentRecord
	if err := db.Model(&assignmentStudentRecord{}).
		Joins("JOIN assignments ON assignments.id = assignment_students.assignment_id").
		Where("assignments.user_id = ?", owner).
		Order("assignment_students.assignment_id, assignment_students.student_id").
		Find(&assignmentScores).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export assignment scores: %w", err)
	}
	for _, row := range assignmentScores {
		dataset.AssignmentScores = append(dataset.AssignmentScores, model.ScoreRow{
			ParentID: row.AssignmentID, StudentID: row.StudentID, Score: row.Score, RecordedAt: row.CompletedAt,
		})
	}

	var exams []examRecord
	if err := db.Where("user_id = ?", owner).Order("exam_date, name").Find(&exams).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export exams: %w", err)
	}
	for _, exam := range exams {
		dataset.Exams = append(dataset.Exams, model.ExamRow{
			ID: exam.ID, SchoolID: exam.SchoolID, AcademicYearID: exam.AcademicYearID, ClassID: exam.ClassID,
			Name: exam.Name, Type: exam.Type, ExamDate: exam.ExamDate, TotalScore: exam.TotalScore,
		})
	}

	var examScores []examStudentRecord
	if err := db.Model(&examStudentRecord{}).
		Joins("JOIN exams ON exams.id = exam_students.exam_id").
		Where("exams.user_id = ?", owner).
		Order("exam_students.exam_id, exam_students.student_id").Find(&examScores).Error; err != nil {
		return model.Dataset{}, fmt.Errorf("export exam scores: %w", err)
	}
	for _, row := range examScores {
		dataset.ExamScores = append(dataset.ExamScores, model.ScoreRow{
			ParentID: row.ExamID, StudentID: row.StudentID, Score: row.Score, RecordedAt: row.MarkedAt,
		})
	}

	return dataset, nil
}

// ReplaceUserData swaps the caller's whole workspace for the dataset. The
// service layer has already rewritten every id, so rows are inserted verbatim.
func (s *Store) ReplaceUserData(ctx context.Context, dataset model.Dataset) error {
	owner := userID(ctx)
	if owner == "" {
		return fmt.Errorf("replace user data: missing user identity")
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := clearOwnedData(tx, owner); err != nil {
			return err
		}
		return insertDataset(tx, owner, dataset)
	})
	return storageError(err)
}

func insertDataset(tx *gorm.DB, owner string, dataset model.Dataset) error {
	for _, school := range dataset.Schools {
		record := schoolRecord{ID: school.ID, UserID: owner, Name: school.Name, IsActive: school.IsActive}
		for position, classroom := range school.Classrooms {
			record.Classrooms = append(record.Classrooms, schoolClassroomRecord{
				SchoolID: school.ID, Classroom: classroom, Position: position,
			})
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("import school %q: %w", school.Name, err)
		}
	}

	segmentsByYear := make(map[string][]academicSegmentRecord, len(dataset.AcademicYears))
	for _, segment := range dataset.AcademicSegments {
		segmentsByYear[segment.AcademicYearID] = append(segmentsByYear[segment.AcademicYearID], academicSegmentRecord{
			ID: segment.ID, AcademicYearID: segment.AcademicYearID, Name: segment.Name, Type: segment.Type,
			StartDate: segment.StartDate, EndDate: segment.EndDate, Position: segment.Position,
			DurationDays: inclusiveDays(segment.StartDate, segment.EndDate),
		})
	}
	for _, year := range dataset.AcademicYears {
		record := academicYearRecord{
			ID: year.ID, UserID: owner, SchoolID: year.SchoolID, Name: year.Name,
			StartDate: year.StartDate, EndDate: year.EndDate, IsCurrent: year.IsCurrent,
			DurationDays: inclusiveDays(year.StartDate, year.EndDate),
			Segments:     segmentsByYear[year.ID],
		}
		if err := tx.Omit("School").Create(&record).Error; err != nil {
			return fmt.Errorf("import academic year %q: %w", year.Name, err)
		}
	}

	for _, student := range dataset.Students {
		record := studentRecord{
			ID: student.ID, UserID: owner, SchoolID: student.SchoolID,
			FirstName: student.FirstName, LastName: student.LastName, Email: student.Email, Phone: student.Phone,
			GuardianName: student.GuardianName, GuardianEmail: student.GuardianEmail, GuardianPhone: student.GuardianPhone,
			ResidentAddress: student.ResidentAddress, PermanentAddress: student.PermanentAddress,
		}
		if err := tx.Omit("School", "Classrooms").Create(&record).Error; err != nil {
			return fmt.Errorf("import student %q: %w", student.FirstName+" "+student.LastName, err)
		}
	}
	for _, classroom := range dataset.StudentClassrooms {
		record := studentClassroomRecord{
			StudentID: classroom.StudentID, AcademicYearID: classroom.AcademicYearID, Classroom: classroom.Classroom,
		}
		if err := tx.Omit("AcademicYear").Create(&record).Error; err != nil {
			return fmt.Errorf("import student classroom: %w", err)
		}
	}
	for _, log := range dataset.StudentLogs {
		record := studentLogRecord{
			ID: log.ID, StudentID: log.StudentID, Kind: log.Kind, Type: log.Type,
			Note: log.Note, CreatedAt: log.CreatedAt,
		}
		if err := tx.Omit("Student").Create(&record).Error; err != nil {
			return fmt.Errorf("import student log: %w", err)
		}
	}

	for _, class := range dataset.Classes {
		record := classRecordFrom(owner, model.Class{
			ID: class.ID, SchoolID: class.SchoolID, AcademicYearID: class.AcademicYearID,
			Name: class.Name, Subject: class.Subject, Description: class.Description,
		})
		if err := tx.Omit("School", "AcademicYear", "Classrooms", "Students").Create(&record).Error; err != nil {
			return fmt.Errorf("import class %q: %w", class.Name, err)
		}
		for _, classroom := range class.Classrooms {
			if err := tx.Create(&classClassroomRecord{ClassID: class.ID, Classroom: classroom}).Error; err != nil {
				return fmt.Errorf("import class classroom: %w", err)
			}
		}
	}
	for _, enrolment := range dataset.ClassStudents {
		record := classStudentRecord{ClassID: enrolment.ClassID, StudentID: enrolment.StudentID}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("import class student: %w", err)
		}
	}

	for _, assignment := range dataset.Assignments {
		record := assignmentRecord{
			ID: assignment.ID, UserID: owner, SchoolID: assignment.SchoolID, AcademicYearID: assignment.AcademicYearID,
			Name: assignment.Name, Type: assignment.Type, DueDate: assignment.DueDate, TotalScore: assignment.TotalScore,
		}
		if assignment.ClassID != "" {
			classID := assignment.ClassID
			record.ClassID = &classID
		}
		if err := tx.Omit("School", "AcademicYear", "Class", "Students").Create(&record).Error; err != nil {
			return fmt.Errorf("import assignment %q: %w", assignment.Name, err)
		}
	}
	for _, score := range dataset.AssignmentScores {
		record := assignmentStudentRecord{
			AssignmentID: score.ParentID, StudentID: score.StudentID,
			Score: score.Score, CompletedAt: score.RecordedAt,
		}
		if err := tx.Omit("Student").Create(&record).Error; err != nil {
			return fmt.Errorf("import assignment score: %w", err)
		}
	}

	for _, exam := range dataset.Exams {
		record := examRecord{
			ID: exam.ID, UserID: owner, SchoolID: exam.SchoolID, AcademicYearID: exam.AcademicYearID,
			ClassID: exam.ClassID, Name: exam.Name, Type: exam.Type, ExamDate: exam.ExamDate, TotalScore: exam.TotalScore,
		}
		if err := tx.Omit("School", "AcademicYear", "Class", "Students").Create(&record).Error; err != nil {
			return fmt.Errorf("import exam %q: %w", exam.Name, err)
		}
	}
	for _, score := range dataset.ExamScores {
		record := examStudentRecord{
			ExamID: score.ParentID, StudentID: score.StudentID,
			Score: score.Score, MarkedAt: score.RecordedAt,
		}
		if err := tx.Omit("Student").Create(&record).Error; err != nil {
			return fmt.Errorf("import exam score: %w", err)
		}
	}
	return nil
}
