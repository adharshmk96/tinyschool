package dataio

import (
	"testing"
	"time"

	"tinyschool-api/internal/model"
)

func sample() model.Dataset {
	score := 18.5
	marked := time.Date(2026, 7, 1, 9, 30, 0, 0, time.UTC)
	return model.Dataset{
		Schools: []model.SchoolRow{{ID: "sch_1", Name: "Tiny School", Classrooms: []string{"8A", "8B"}, IsActive: true}},
		AcademicYears: []model.AcademicYearRow{
			{ID: "ay_1", SchoolID: "sch_1", Name: "2026-27", StartDate: "2026-06-01", EndDate: "2027-03-31", IsCurrent: true},
		},
		AcademicSegments: []model.AcademicSegmentRow{
			{ID: "seg_1", AcademicYearID: "ay_1", Name: "Term 1", Type: "term", StartDate: "2026-06-01", EndDate: "2026-09-30"},
		},
		Students: []model.StudentRow{
			{ID: "stu_1", SchoolID: "sch_1", FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"},
		},
		StudentClassrooms: []model.StudentClassroomRow{{StudentID: "stu_1", AcademicYearID: "ay_1", Classroom: "8A"}},
		StudentLogs: []model.StudentLogRow{
			{ID: "note_1", StudentID: "stu_1", Kind: "note", Note: "Doing well", CreatedAt: marked},
		},
		Classes: []model.ClassRow{
			{ID: "cls_1", SchoolID: "sch_1", AcademicYearID: "ay_1", Name: "Maths 8", Subject: "Maths", Classrooms: []string{"8A", "8B"}},
		},
		ClassStudents: []model.ClassStudentRow{{ClassID: "cls_1", StudentID: "stu_1"}},
		Exams: []model.ExamRow{
			{ID: "exam_1", SchoolID: "sch_1", AcademicYearID: "ay_1", ClassID: "cls_1", Name: "Midterm", Type: "written", ExamDate: "2026-09-01", TotalScore: 20},
		},
		ExamScores: []model.ScoreRow{{ParentID: "exam_1", StudentID: "stu_1", Score: &score, RecordedAt: &marked}},
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	for _, format := range []struct {
		format   Format
		filename string
	}{
		{FormatXLSX, "export.xlsx"},
		{FormatCSV, "export.zip"},
	} {
		t.Run(string(format.format), func(t *testing.T) {
			encoded, err := Encode(sample(), format.format)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := Decode(format.filename, encoded)
			if err != nil {
				t.Fatal(err)
			}
			original := sample()
			if len(decoded.Schools) != 1 || decoded.Schools[0].Name != "Tiny School" {
				t.Fatalf("schools = %+v", decoded.Schools)
			}
			if got, want := decoded.Schools[0].Classrooms, original.Schools[0].Classrooms; len(got) != len(want) || got[1] != want[1] {
				t.Fatalf("classrooms = %v, want %v", got, want)
			}
			if !decoded.AcademicYears[0].IsCurrent {
				t.Fatal("isCurrent was lost")
			}
			if decoded.StudentLogs[0].CreatedAt.UTC() != original.StudentLogs[0].CreatedAt {
				t.Fatalf("createdAt = %v", decoded.StudentLogs[0].CreatedAt)
			}
			score := decoded.ExamScores[0].Score
			if score == nil || *score != 18.5 {
				t.Fatalf("score = %v", score)
			}
			if decoded.ExamScores[0].RecordedAt == nil {
				t.Fatal("markedAt was lost")
			}
			if decoded.Exams[0].TotalScore != 20 {
				t.Fatalf("totalScore = %v", decoded.Exams[0].TotalScore)
			}
		})
	}
}

// A hand-written sheet may reorder columns, retitle them, drop optional ones
// and leave blank lines behind; all of that must still import.
func TestDecodeToleratesHandEditedCSV(t *testing.T) {
	file := "Last Name,FirstName,ID,SchoolId\n\nLovelace,Ada,stu_1,sch_1\n"
	decoded, err := Decode("Students.csv", []byte(file))
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Students) != 1 {
		t.Fatalf("students = %+v", decoded.Students)
	}
	if decoded.Students[0].ID != "stu_1" || decoded.Students[0].FirstName != "Ada" {
		t.Fatalf("student = %+v", decoded.Students[0])
	}
	// "Last Name" is not a column of the sheet, so it is simply ignored.
	if decoded.Students[0].LastName != "" {
		t.Fatalf("lastName = %q", decoded.Students[0].LastName)
	}
}

func TestDecodeReportsBadCells(t *testing.T) {
	file := "id,schoolId,academicYearId,classId,name,examDate,totalScore\nexam_1,sch_1,ay_1,cls_1,Midterm,2026-09-01,twenty\n"
	if _, err := Decode("exams.csv", []byte(file)); err == nil {
		t.Fatal("expected an error for a non-numeric totalScore")
	} else if got := err.Error(); got != "exams row 2: totalScore must be a number" {
		t.Fatalf("error = %q", got)
	}
}

func TestDecodeRejectsUnknownSheet(t *testing.T) {
	if _, err := Decode("teachers.csv", []byte("id,name\nt_1,Alex\n")); err == nil {
		t.Fatal("expected an error for an unknown sheet name")
	}
}
