package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/jung-kurt/gofpdf"
)

// BuildTranscriptPDF generates a PDF from a StudentTranscript and returns the bytes.
func BuildTranscriptPDF(transcript *entity.StudentTranscript) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "MYRMEX University", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 13)
	pdf.CellFormat(0, 8, "Student Academic Transcript", "", 1, "C", false, 0, "")
	pdf.Ln(4)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(4)

	// Student info
	student := transcript.Student
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(30, 6, "Student:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(80, 6, student.FullName, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(25, 6, "Code:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(40, 6, student.StudentCode, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(30, 6, "Email:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(80, 6, student.Email, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(25, 6, "Year:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(40, 6, fmt.Sprintf("%d", student.EnrollmentYear), "", 1, "L", false, 0, "")

	pdf.Ln(3)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(4)

	// Group entries by semester
	type semesterGroup struct {
		semesterID string
		entries    []entity.TranscriptEntry
	}
	seen := make(map[string]int)
	var groups []semesterGroup
	for _, e := range transcript.Entries {
		idx, ok := seen[e.SemesterID]
		if !ok {
			seen[e.SemesterID] = len(groups)
			groups = append(groups, semesterGroup{semesterID: e.SemesterID, entries: []entity.TranscriptEntry{e}})
		} else {
			groups[idx].entries = append(groups[idx].entries, e)
		}
	}

	// Table header helper
	drawTableHeader := func() {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(28, 7, "Code", "1", 0, "C", true, 0, "")
		pdf.CellFormat(72, 7, "Subject", "1", 0, "C", true, 0, "")
		pdf.CellFormat(18, 7, "Credits", "1", 0, "C", true, 0, "")
		pdf.CellFormat(22, 7, "Grade", "1", 0, "C", true, 0, "")
		pdf.CellFormat(18, 7, "Letter", "1", 0, "C", true, 0, "")
		pdf.CellFormat(22, 7, "Status", "1", 1, "C", true, 0, "")
		pdf.SetFillColor(255, 255, 255)
	}

	for _, group := range groups {
		// Semester label
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(0, 7, fmt.Sprintf("Semester: %s", group.semesterID[:8]+"..."), "", 1, "L", false, 0, "")

		drawTableHeader()

		var semCredits int32
		var semWeight float64
		for _, e := range group.entries {
			pdf.SetFont("Helvetica", "", 9)
			subjectCode := truncate(e.SubjectCode, 10)
			subjectName := truncate(e.SubjectName, 35)
			grade := ""
			if e.GradeNumeric > 0 {
				grade = fmt.Sprintf("%.2f", e.GradeNumeric)
			}
			pdf.CellFormat(28, 6, subjectCode, "1", 0, "C", false, 0, "")
			pdf.CellFormat(72, 6, subjectName, "1", 0, "L", false, 0, "")
			pdf.CellFormat(18, 6, fmt.Sprintf("%d", e.Credits), "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 6, grade, "1", 0, "C", false, 0, "")
			pdf.CellFormat(18, 6, e.GradeLetter, "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 6, e.Status, "1", 1, "C", false, 0, "")

			semCredits += e.Credits
			if e.GradeNumeric > 0 {
				semWeight += e.GradeNumeric * float64(e.Credits)
			}
		}

		// Semester summary
		semGPA := 0.0
		if semCredits > 0 {
			semGPA = semWeight / float64(semCredits)
		}
		pdf.SetFont("Helvetica", "I", 9)
		pdf.CellFormat(0, 5, fmt.Sprintf("Semester GPA: %.2f  |  Credits: %d", semGPA, semCredits), "", 1, "R", false, 0, "")
		pdf.Ln(3)
	}

	// Overall summary
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(3)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 7, fmt.Sprintf(
		"Overall GPA: %.2f  |  Total Credits: %d  |  Passed Credits: %d",
		transcript.GPA, transcript.TotalCredits, transcript.PassedCredits,
	), "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "I", 8)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")), "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
