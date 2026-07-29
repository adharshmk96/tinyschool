package gormsqlite

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (s *Store) Seed(ctx context.Context) error {
	var count int64
	for _, table := range []string{"users", "schools", "academic_years", "students", "classes", "assignments", "exams"} {
		var tableCount int64
		if err := s.db.WithContext(ctx).Table(table).Count(&tableCount).Error; err != nil {
			return fmt.Errorf("check seed table %s: %w", table, err)
		}
		count += tableCount
	}
	if count != 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Transaction(seedFixtures); err != nil {
		return fmt.Errorf("seed fixtures: %w", err)
	}
	return nil
}

func seedFixtures(tx *gorm.DB) error {
	if err := tx.Create(&userRecord{
		ID: "usr_001", Name: "Alex Morgan", Email: "alex@tinyschool.local",
		PasswordHash: "$2a$10$WYpbkVsCeojlhkpsUuk4de4DWZSZpCyQyH2FPQ9PLE7WSODy95HNy",
	}).Error; err != nil {
		return err
	}
	schools := []schoolRecord{
		{ID: "sch_001", UserID: "usr_001", Name: "Tiny School Academy", IsActive: true},
		{ID: "sch_002", UserID: "usr_001", Name: "Tiny School Primary", IsActive: false},
	}
	if err := tx.Create(&schools).Error; err != nil {
		return err
	}
	for schoolID, classrooms := range map[string][]string{
		"sch_001": {"8A", "8B", "8C", "9A", "9B", "9C", "10A", "10B", "10C"},
		"sch_002": {"8A", "8B", "8C", "9A", "9B", "9C", "10A", "10B", "10C"},
	} {
		for position, classroom := range classrooms {
			if err := tx.Create(&schoolClassroomRecord{SchoolID: schoolID, Classroom: classroom, Position: position}).Error; err != nil {
				return err
			}
		}
	}
	years := []academicYearRecord{
		{ID: "ay_2026", UserID: "usr_001", SchoolID: "sch_001", Name: "2026–27", StartDate: "2026-06-01", EndDate: "2027-03-31", DurationDays: 304, IsCurrent: true},
		{ID: "ay_2025", UserID: "usr_001", SchoolID: "sch_001", Name: "2025–26", StartDate: "2025-06-02", EndDate: "2026-03-31", DurationDays: 303},
	}
	if err := tx.Create(&years).Error; err != nil {
		return err
	}
	segments := []academicSegmentRecord{
		{ID: "seg_001", AcademicYearID: "ay_2026", Name: "Term 1", Type: "term", DurationDays: 120, StartDate: "2026-06-01", EndDate: "2026-09-28", Position: 0},
		{ID: "seg_002", AcademicYearID: "ay_2026", Name: "Autumn Break", Type: "vacation", DurationDays: 10, StartDate: "2026-09-29", EndDate: "2026-10-08", Position: 1},
		{ID: "seg_003", AcademicYearID: "ay_2026", Name: "Term 2", Type: "term", DurationDays: 174, StartDate: "2026-10-09", EndDate: "2027-03-31", Position: 2},
		{ID: "seg_004", AcademicYearID: "ay_2025", Name: "Term 1", Type: "term", DurationDays: 151, StartDate: "2025-06-02", EndDate: "2025-10-30", Position: 0},
		{ID: "seg_005", AcademicYearID: "ay_2025", Name: "Term 2", Type: "term", DurationDays: 152, StartDate: "2025-10-31", EndDate: "2026-03-31", Position: 1},
	}
	if err := tx.Create(&segments).Error; err != nil {
		return err
	}
	students := []studentRecord{
		studentFixture("stu_001", "Maya", "Patel", "8A", "maya.patel@example.test", "+91 98765 11001", "Rina Patel"),
		studentFixture("stu_002", "Noah", "Williams", "8A", "noah.williams@example.test", "+91 98765 11002", "Sophie Williams"),
		studentFixture("stu_003", "Aarav", "Shah", "8B", "aarav.shah@example.test", "+91 98765 11003", "Neel Shah"),
		studentFixture("stu_004", "Emma", "Chen", "9A", "emma.chen@example.test", "+91 98765 11004", "Wei Chen"),
		studentFixture("stu_005", "Liam", "Brown", "9B", "liam.brown@example.test", "+91 98765 11005", "Amelia Brown"),
		studentFixture("stu_006", "Olivia", "Martin", "10A", "olivia.martin@example.test", "+91 98765 11006", "Lucas Martin"),
		studentFixture("stu_007", "Ethan", "Wilson", "10A", "ethan.wilson@example.test", "+91 98765 11007", "Grace Wilson"),
		studentFixture("stu_008", "Sophia", "Garcia", "10B", "sophia.garcia@example.test", "+91 98765 11008", "Mateo Garcia"),
	}
	if err := tx.Omit("Classrooms").Create(&students).Error; err != nil {
		return err
	}
	if err := seedStudentClassrooms(tx, students); err != nil {
		return err
	}
	classes := []classRecord{
		{ID: "cls_math8", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Grade 8 Mathematics", Subject: "Mathematics", Description: "Core mathematics with an emphasis on algebra and geometry."},
		{ID: "cls_sci8", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Grade 8 Science", Subject: "Science", Description: "Hands-on life and physical sciences."},
		{ID: "cls_eng9", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Grade 9 English", Subject: "English", Description: "Reading comprehension and creative writing."},
		{ID: "cls_hist10", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Grade 10 History", Subject: "History", Description: "World history through primary sources."},
	}
	if err := tx.Omit("Classrooms", "Students").Create(&classes).Error; err != nil {
		return err
	}
	for classID, rooms := range map[string][]string{
		"cls_math8":  {"8A", "8B"},
		"cls_sci8":   {"8A", "8B", "8C"},
		"cls_eng9":   {"9A", "9B"},
		"cls_hist10": {"10A", "10B"},
	} {
		for _, classroom := range rooms {
			if err := tx.Create(&classClassroomRecord{ClassID: classID, Classroom: classroom}).Error; err != nil {
				return err
			}
		}
	}
	for classID, ids := range map[string][]string{
		"cls_math8": {"stu_001", "stu_002", "stu_003"}, "cls_sci8": {"stu_001", "stu_002", "stu_003"},
		"cls_eng9": {"stu_004", "stu_005"}, "cls_hist10": {"stu_006", "stu_007", "stu_008"},
	} {
		for _, id := range ids {
			if err := tx.Create(&classStudentRecord{ClassID: classID, StudentID: id}).Error; err != nil {
				return err
			}
		}
	}
	if err := seedAssignments(tx); err != nil {
		return err
	}
	if err := seedExams(tx); err != nil {
		return err
	}
	logs := []studentLogRecord{
		{ID: "beh_001", StudentID: "stu_001", Kind: "behaviour", Type: "positive", Note: "Helped a classmate during lab work.", CreatedAt: mustTime("2026-07-24T10:30:00Z")},
		{ID: "note_001", StudentID: "stu_001", Kind: "note", Note: "Strong improvement in algebra this month.", CreatedAt: mustTime("2026-07-25T08:15:00Z")},
		{ID: "beh_002", StudentID: "stu_002", Kind: "behaviour", Type: "need_attention", Note: "Two assignments submitted late.", CreatedAt: mustTime("2026-07-23T09:00:00Z")},
		{ID: "note_002", StudentID: "stu_002", Kind: "note", Note: "Schedule a guardian check-in.", CreatedAt: mustTime("2026-07-24T12:00:00Z")},
		{ID: "beh_006_1", StudentID: "stu_006", Kind: "behaviour", Type: "positive", Note: "Led the primary-source discussion with thoughtful questions.", CreatedAt: mustTime("2026-07-26T09:35:00Z")},
		{ID: "beh_006_2", StudentID: "stu_006", Kind: "behaviour", Type: "need_attention", Note: "Needs a reminder to bring the history workbook.", CreatedAt: mustTime("2026-07-22T11:10:00Z")},
		{ID: "note_006_1", StudentID: "stu_006", Kind: "note", Note: "Olivia responds well to visual timelines and source maps.", CreatedAt: mustTime("2026-07-25T13:20:00Z")},
		{ID: "note_006_2", StudentID: "stu_006", Kind: "note", Note: "Guardian requested a progress update after the next assessment.", CreatedAt: mustTime("2026-07-20T08:45:00Z")},
	}
	return tx.Create(&logs).Error
}

func studentFixture(id, first, last, classroom, email, phone, guardian string) studentRecord {
	return studentRecord{
		ID: id, UserID: "usr_001", SchoolID: "sch_001", FirstName: first, LastName: last,
		Email: email, Phone: phone, GuardianName: guardian,
		Classrooms: []studentClassroomRecord{{StudentID: id, AcademicYearID: "ay_2026", Classroom: classroom}},
	}
}

// previousClassroom steps a "NA" label one year back for the prior year rows.
func previousClassroom(classroom string) string {
	var number int
	var division string
	if _, err := fmt.Sscanf(classroom, "%d%s", &number, &division); err != nil || number <= 1 {
		return classroom
	}
	return fmt.Sprintf("%d%s", number-1, division)
}

// seedStudentClassrooms records each seeded student in the current year with the
// classroom from their fixture, and one year lower in the previous year.
func seedStudentClassrooms(tx *gorm.DB, students []studentRecord) error {
	for _, student := range students {
		classroom := student.Classrooms[0].Classroom
		rows := []studentClassroomRecord{
			{StudentID: student.ID, AcademicYearID: "ay_2026", Classroom: classroom},
			{StudentID: student.ID, AcademicYearID: "ay_2025", Classroom: previousClassroom(classroom)},
		}
		for _, row := range rows {
			if err := tx.Omit("AcademicYear").Create(&row).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedAssignments(tx *gorm.DB) error {
	classIDs := map[string]*string{}
	for _, value := range []string{"cls_math8", "cls_sci8", "cls_eng9", "cls_hist10"} {
		id := value
		classIDs[value] = &id
	}
	records := []assignmentRecord{
		{ID: "asg_001", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Algebra Practice", Type: "class", DueDate: "2026-08-02", TotalScore: 20, ClassID: classIDs["cls_math8"]},
		{ID: "asg_002", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Plant Cell Model", Type: "class", DueDate: "2026-08-08", TotalScore: 30, ClassID: classIDs["cls_sci8"]},
		{ID: "asg_003", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Short Story Draft", Type: "class", DueDate: "2026-08-10", TotalScore: 25, ClassID: classIDs["cls_eng9"]},
		{ID: "asg_004", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Geometry Project", Type: "individual", DueDate: "2026-08-18", TotalScore: 50},
		{ID: "asg_005", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Ancient Civilizations Essay", Type: "class", DueDate: "2026-08-22", TotalScore: 40, ClassID: classIDs["cls_hist10"]},
		{ID: "asg_006", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", Name: "Reading Reflection", Type: "individual", DueDate: "2026-07-30", TotalScore: 10},
	}
	if err := tx.Create(&records).Error; err != nil {
		return err
	}
	rosters := map[string][]string{"asg_001": {"stu_001", "stu_002", "stu_003"}, "asg_002": {"stu_001", "stu_002", "stu_003"}, "asg_003": {"stu_004", "stu_005"}, "asg_004": {"stu_001"}, "asg_005": {"stu_006", "stu_007", "stu_008"}, "asg_006": {"stu_005"}}
	for assignmentID, ids := range rosters {
		for _, id := range ids {
			row := assignmentStudentRecord{AssignmentID: assignmentID, StudentID: id}
			switch assignmentID + "/" + id {
			case "asg_001/stu_001":
				row.Score, row.CompletedAt = number(18), instant("2026-07-27T09:10:00Z")
			case "asg_001/stu_003":
				row.Score, row.CompletedAt = number(16), instant("2026-07-28T08:20:00Z")
			case "asg_002/stu_003":
				row.Score, row.CompletedAt = number(27), instant("2026-07-26T13:00:00Z")
			case "asg_005/stu_006":
				row.Score, row.CompletedAt = number(34), instant("2026-08-20T14:05:00Z")
			case "asg_005/stu_008":
				row.Score, row.CompletedAt = number(36), instant("2026-08-21T10:40:00Z")
			case "asg_006/stu_005":
				row.Score, row.CompletedAt = number(9), instant("2026-07-28T15:00:00Z")
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedExams(tx *gorm.DB) error {
	records := []examRecord{
		{ID: "exam_001", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", ClassID: "cls_math8", Name: "Mathematics Midterm", Type: "midterm", ExamDate: "2026-08-15", TotalScore: 100},
		{ID: "exam_002", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", ClassID: "cls_sci8", Name: "Science Quiz", Type: "quiz", ExamDate: "2026-08-20", TotalScore: 30},
		{ID: "exam_003", UserID: "usr_001", SchoolID: "sch_001", AcademicYearID: "ay_2026", ClassID: "cls_eng9", Name: "English Assessment", Type: "assessment", ExamDate: "2026-08-24", TotalScore: 50},
	}
	if err := tx.Create(&records).Error; err != nil {
		return err
	}
	for examID, ids := range map[string][]string{"exam_001": {"stu_001", "stu_002", "stu_003"}, "exam_002": {"stu_001", "stu_002", "stu_003"}, "exam_003": {"stu_004", "stu_005"}} {
		for _, id := range ids {
			row := examStudentRecord{ExamID: examID, StudentID: id}
			switch examID + "/" + id {
			case "exam_001/stu_001":
				row.Score, row.MarkedAt = number(86), instant("2026-08-15T11:00:00Z")
			case "exam_001/stu_002":
				row.Score, row.MarkedAt = number(74), instant("2026-08-15T11:05:00Z")
			case "exam_003/stu_004":
				row.Score, row.MarkedAt = number(44), instant("2026-08-24T10:30:00Z")
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func instant(value string) *time.Time { result := mustTime(value); return &result }
func number(value float64) *float64   { return &value }
