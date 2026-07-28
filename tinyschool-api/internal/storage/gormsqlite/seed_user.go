package gormsqlite

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SeedUserData fills the account owned by email with a demo data set: one
// active school and current academic year (reused when they already exist),
// thirty students, five classes, four assignments per class, and four exams
// per class. Assignment and exam results are included to make dashboards and
// performance views representative. It is only run from the `seed` command.
func (s *Store) SeedUserData(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	var user userRecord
	if err := s.db.WithContext(ctx).First(&user, "email = ? COLLATE NOCASE", email).Error; err != nil {
		return fmt.Errorf("find user %s: %w", email, storageError(err))
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return seedUserFixtures(tx, user.ID)
	})
}

var (
	seedGrades = []string{"Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5", "Grade 6", "Grade 7", "Grade 8", "Grade 9", "Grade 10"}
	seedNames  = [][2]string{
		{"Maya", "Patel"}, {"Noah", "Williams"}, {"Aarav", "Shah"}, {"Emma", "Chen"}, {"Liam", "Brown"},
		{"Olivia", "Martin"}, {"Ethan", "Wilson"}, {"Sophia", "Garcia"}, {"Ravi", "Nair"}, {"Isla", "Fernandes"},
		{"Arjun", "Mehta"}, {"Ava", "Thompson"}, {"Kabir", "Kapoor"}, {"Mia", "Rodriguez"}, {"Leo", "Anderson"},
		{"Ananya", "Rao"}, {"Lucas", "Taylor"}, {"Zara", "Khan"}, {"Henry", "Moore"}, {"Diya", "Iyer"},
		{"Jack", "Harris"}, {"Meera", "Joshi"}, {"Oliver", "Clark"}, {"Ishaan", "Gupta"}, {"Ella", "Lewis"},
		{"Vihaan", "Singh"}, {"Amelia", "Walker"}, {"Aditya", "Desai"}, {"Grace", "Hall"}, {"Nina", "Thomas"},
	}
	seedSubjects = []struct{ subject, grade string }{
		{"Mathematics", "Grade 6"}, {"Science", "Grade 6"}, {"English", "Grade 7"}, {"History", "Grade 7"}, {"Geography", "Grade 8"},
	}
	seedAssignmentNames = []string{"Practice Set", "Group Project", "Worksheet", "Research Task"}
	seedExamSpecs       = []struct{ name, kind string }{
		{"Unit Test", "quiz"}, {"Midterm", "midterm"}, {"Assessment", "assessment"}, {"Final", "final"},
	}
)

func seedUserFixtures(tx *gorm.DB, ownerID string) error {
	school, err := seedSchool(tx, ownerID)
	if err != nil {
		return err
	}
	year, err := seedAcademicYear(tx, ownerID, school.ID)
	if err != nil {
		return err
	}

	students := make([]studentRecord, 0, len(seedNames))
	for i, name := range seedNames {
		id := seedID("stu")
		student := studentRecord{
			ID: id, UserID: ownerID, SchoolID: school.ID,
			FirstName: name[0], LastName: name[1],
			Email:        fmt.Sprintf("%s.%s@example.test", strings.ToLower(name[0]), strings.ToLower(name[1])),
			Phone:        fmt.Sprintf("+91 98765 %05d", 11001+i),
			GuardianName: "Guardian of " + name[0],
		}
		if err := tx.Omit("Grades").Create(&student).Error; err != nil {
			return fmt.Errorf("create student: %w", err)
		}
		// Students are enrolled six per class below, so give each the grade of
		// the class they will land in.
		grade := seedSubjects[i/6].grade
		if err := tx.Omit("AcademicYear").Create(&studentGradeRecord{StudentID: id, AcademicYearID: year.ID, Grade: grade}).Error; err != nil {
			return fmt.Errorf("create student grade: %w", err)
		}
		students = append(students, student)
	}

	for i, spec := range seedSubjects {
		classID := seedID("cls")
		class := classRecord{
			ID: classID, UserID: ownerID, SchoolID: school.ID, AcademicYearID: year.ID,
			Name:        fmt.Sprintf("%s %s", spec.subject, spec.grade),
			Subject:     spec.subject,
			Grade:       spec.grade,
			Description: fmt.Sprintf("Seeded %s class for %s.", strings.ToLower(spec.subject), spec.grade),
		}
		if err := tx.Omit("Students").Create(&class).Error; err != nil {
			return fmt.Errorf("create class: %w", err)
		}
		// Six students per class, so all thirty are enrolled exactly once.
		roster := students[i*6 : i*6+6]
		for _, student := range roster {
			if err := tx.Create(&classStudentRecord{ClassID: classID, StudentID: student.ID}).Error; err != nil {
				return fmt.Errorf("enrol student: %w", err)
			}
		}
		if err := seedClassAssignments(tx, ownerID, school.ID, year.ID, class, roster); err != nil {
			return err
		}
		if err := seedClassExams(tx, ownerID, school.ID, year.ID, class, roster); err != nil {
			return err
		}
	}
	return nil
}

func seedClassAssignments(tx *gorm.DB, ownerID, schoolID, yearID string, class classRecord, roster []studentRecord) error {
	for i, name := range seedAssignmentNames {
		id := seedID("asg")
		classID := class.ID
		record := assignmentRecord{
			ID: id, UserID: ownerID, SchoolID: schoolID, AcademicYearID: yearID,
			Name:       fmt.Sprintf("%s %s %d", class.Subject, name, i+1),
			Type:       "class",
			DueDate:    seedDate(7 * (i + 1)),
			TotalScore: 20 + float64(i)*10,
			ClassID:    &classID,
		}
		if err := tx.Omit("Students").Create(&record).Error; err != nil {
			return fmt.Errorf("create assignment: %w", err)
		}
		for studentIndex, student := range roster {
			row := assignmentStudentRecord{AssignmentID: id, StudentID: student.ID}
			// Leave some work pending while producing a broad, believable score
			// distribution for completed work.
			if (i+studentIndex)%4 != 0 {
				score := float64(58+((i*11+studentIndex*7)%39)) * record.TotalScore / 100
				completedAt := time.Now().UTC().AddDate(0, 0, -(i + studentIndex + 1))
				row.Score, row.CompletedAt = &score, &completedAt
			}
			if err := tx.Create(&row).Error; err != nil {
				return fmt.Errorf("assign student: %w", err)
			}
		}
	}
	return nil
}

func seedClassExams(tx *gorm.DB, ownerID, schoolID, yearID string, class classRecord, roster []studentRecord) error {
	for i, spec := range seedExamSpecs {
		id := seedID("exam")
		record := examRecord{
			ID: id, UserID: ownerID, SchoolID: schoolID, AcademicYearID: yearID, ClassID: class.ID,
			Name:       fmt.Sprintf("%s %s", class.Subject, spec.name),
			Type:       spec.kind,
			ExamDate:   seedDate(14 * (i + 1)),
			TotalScore: 50 + float64(i)*25,
		}
		if err := tx.Omit("Students").Create(&record).Error; err != nil {
			return fmt.Errorf("create exam: %w", err)
		}
		for studentIndex, student := range roster {
			row := examStudentRecord{ExamID: id, StudentID: student.ID}
			if i < 2 || studentIndex%3 != 0 {
				score := float64(55+((i*13+studentIndex*9)%42)) * record.TotalScore / 100
				markedAt := time.Now().UTC().AddDate(0, 0, -(i*7 + studentIndex + 1))
				row.Score, row.MarkedAt = &score, &markedAt
			}
			if err := tx.Create(&row).Error; err != nil {
				return fmt.Errorf("add exam student: %w", err)
			}
		}
	}
	return nil
}

// seedSchool returns the account's active school, creating one when the
// account has no school yet.
func seedSchool(tx *gorm.DB, ownerID string) (schoolRecord, error) {
	var existing schoolRecord
	err := tx.Where("user_id = ?", ownerID).Order("is_active DESC, id").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !isNotFound(err) {
		return schoolRecord{}, fmt.Errorf("find school: %w", err)
	}
	school := schoolRecord{ID: seedID("sch"), UserID: ownerID, Name: "Tiny School Academy", IsActive: true}
	for position, grade := range seedGrades {
		school.Grades = append(school.Grades, schoolGradeRecord{SchoolID: school.ID, Grade: grade, Position: position})
	}
	if err := tx.Create(&school).Error; err != nil {
		return schoolRecord{}, fmt.Errorf("create school: %w", err)
	}
	return school, nil
}

// seedAcademicYear returns the school's current academic year, creating one
// spanning the current June–March session when none exists.
func seedAcademicYear(tx *gorm.DB, ownerID, schoolID string) (academicYearRecord, error) {
	var existing academicYearRecord
	err := tx.Where("user_id = ? AND school_id = ?", ownerID, schoolID).Order("is_current DESC, start_date DESC").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !isNotFound(err) {
		return academicYearRecord{}, fmt.Errorf("find academic year: %w", err)
	}
	now := time.Now().UTC()
	startYear := now.Year()
	if now.Month() < time.June {
		startYear--
	}
	start := time.Date(startYear, time.June, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(startYear+1, time.March, 31, 0, 0, 0, 0, time.UTC)
	middle := time.Date(startYear, time.October, 8, 0, 0, 0, 0, time.UTC)
	year := academicYearRecord{
		ID: seedID("ay"), UserID: ownerID, SchoolID: schoolID,
		Name:      fmt.Sprintf("%d–%02d", startYear, (startYear+1)%100),
		StartDate: day(start), EndDate: day(end),
		DurationDays: days(start, end), IsCurrent: true,
	}
	year.Segments = []academicSegmentRecord{
		{ID: seedID("seg"), AcademicYearID: year.ID, Name: "Term 1", Type: "term", DurationDays: days(start, middle), StartDate: day(start), EndDate: day(middle), Position: 0},
		{ID: seedID("seg"), AcademicYearID: year.ID, Name: "Term 2", Type: "term", DurationDays: days(middle.AddDate(0, 0, 1), end), StartDate: day(middle.AddDate(0, 0, 1)), EndDate: day(end), Position: 1},
	}
	if err := tx.Model(&academicYearRecord{}).Where("user_id = ? AND school_id = ?", ownerID, schoolID).Update("is_current", false).Error; err != nil {
		return academicYearRecord{}, fmt.Errorf("clear current year: %w", err)
	}
	if err := tx.Create(&year).Error; err != nil {
		return academicYearRecord{}, fmt.Errorf("create academic year: %w", err)
	}
	return year, nil
}

func isNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

func day(value time.Time) string { return value.Format("2006-01-02") }

func days(start, end time.Time) int { return int(end.Sub(start).Hours()/24) + 1 }

func seedDate(offset int) string { return day(time.Now().UTC().AddDate(0, 0, offset)) }

func seedID(prefix string) string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("generate %s id: %v", prefix, err))
	}
	return prefix + "_" + hex.EncodeToString(bytes)
}
