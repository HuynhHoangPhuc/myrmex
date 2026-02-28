package export

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// PDFGenerator creates PDF reports from analytics data.
type PDFGenerator struct {
	repo *persistence.AnalyticsRepository
}

// NewPDFGenerator constructs a PDFGenerator with the given repository.
func NewPDFGenerator(repo *persistence.AnalyticsRepository) *PDFGenerator {
	return &PDFGenerator{repo: repo}
}

// GenerateWorkloadReport returns a PDF with teacher workload data for the given semester.
// Columns: Teacher | Department | Subject | Hours/Week | Total Hours
func (g *PDFGenerator) GenerateWorkloadReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	stats, err := g.repo.GetWorkloadStats(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch workload stats: %w", err)
	}

	title := "Teacher Workload Report - " + g.semesterLabel(ctx, sid)
	pdf := newPDF(title)
	headers := []string{"Teacher", "Department", "Subject", "Hours/Week", "Total Hours"}
	widths := []float64{60, 55, 40, 32, 33}
	tbl := newTableRenderer(pdf, widths)
	tbl.header(headers)

	for _, s := range stats {
		tbl.row([]string{
			s.TeacherName,
			s.DepartmentName,
			s.SubjectCode,
			fmt.Sprintf("%.1f", s.HoursPerWeek),
			fmt.Sprintf("%.1f", s.TotalHours),
		})
	}

	return pdfBytes(pdf)
}

// GenerateUtilizationReport returns a PDF with department utilization data for the given semester.
// Columns: Department | Assigned Slots | Total Slots | Utilization %
func (g *PDFGenerator) GenerateUtilizationReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	stats, err := g.repo.GetUtilizationStats(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch utilization stats: %w", err)
	}

	title := "Department Utilization Report - " + g.semesterLabel(ctx, sid)
	pdf := newPDF(title)
	headers := []string{"Department", "Assigned Slots", "Total Slots", "Utilization %"}
	widths := []float64{80, 45, 40, 50}
	tbl := newTableRenderer(pdf, widths)
	tbl.header(headers)

	for _, s := range stats {
		tbl.row([]string{
			s.DepartmentName,
			fmt.Sprintf("%d", s.AssignedSlots),
			fmt.Sprintf("%d", s.TotalSlots),
			fmt.Sprintf("%.1f%%", s.UtilizationPct),
		})
	}

	return pdfBytes(pdf)
}

// GenerateScheduleReport returns a PDF with schedule metrics for the given semester.
// Columns: Semester | Assigned Slots | Total Slots | Fill Rate %
func (g *PDFGenerator) GenerateScheduleReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	metrics, err := g.repo.GetScheduleMetrics(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch schedule metrics: %w", err)
	}

	title := "Schedule Metrics Report - " + g.semesterLabel(ctx, sid)
	pdf := newPDF(title)
	headers := []string{"Semester", "Assigned Slots", "Total Slots", "Fill Rate %"}
	widths := []float64{80, 45, 40, 50}
	tbl := newTableRenderer(pdf, widths)
	tbl.header(headers)

	for _, m := range metrics {
		fillRate := 0.0
		if m.TotalSlots > 0 {
			fillRate = float64(m.AssignedSlots) / float64(m.TotalSlots) * 100
		}
		tbl.row([]string{
			m.SemesterName,
			fmt.Sprintf("%d", m.AssignedSlots),
			fmt.Sprintf("%d", m.TotalSlots),
			fmt.Sprintf("%.1f%%", fillRate),
		})
	}

	return pdfBytes(pdf)
}

// semesterLabel returns the semester name for the title, or "All Semesters" when no filter is set.
func (g *PDFGenerator) semesterLabel(ctx context.Context, sid uuid.UUID) string {
	if sid == uuid.Nil {
		return "All Semesters"
	}
	name, _ := g.repo.GetSemesterName(ctx, sid)
	if name == "" {
		return "Unknown Semester"
	}
	return name
}

// tableRenderer draws a centered, alternating-row PDF table.
type tableRenderer struct {
	pdf    *gofpdf.Fpdf
	startX float64
	widths []float64
	rowIdx int
}

// newTableRenderer creates a renderer that horizontally centers the table on the page.
func newTableRenderer(pdf *gofpdf.Fpdf, widths []float64) *tableRenderer {
	pageW, _ := pdf.GetPageSize()
	total := 0.0
	for _, w := range widths {
		total += w
	}
	return &tableRenderer{
		pdf:    pdf,
		startX: (pageW - total) / 2,
		widths: widths,
	}
}

// header writes a bold blue header row.
func (t *tableRenderer) header(headers []string) {
	t.pdf.SetFont("Arial", "B", 10)
	t.pdf.SetFillColor(66, 135, 245)
	t.pdf.SetTextColor(255, 255, 255)
	t.pdf.SetX(t.startX)
	for i, h := range headers {
		t.pdf.CellFormat(t.widths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	t.pdf.Ln(-1)
	t.pdf.SetTextColor(0, 0, 0)
}

// row writes a data row with alternating background colors.
func (t *tableRenderer) row(cells []string) {
	t.pdf.SetFont("Arial", "", 9)
	if t.rowIdx%2 == 0 {
		t.pdf.SetFillColor(245, 248, 255)
	} else {
		t.pdf.SetFillColor(255, 255, 255)
	}
	t.rowIdx++
	t.pdf.SetX(t.startX)
	for i, cell := range cells {
		t.pdf.CellFormat(t.widths[i], 7, cell, "1", 0, "L", true, 0, "")
	}
	t.pdf.Ln(-1)
}

// newPDF creates a new landscape A4 PDF with a centred title.
func newPDF(title string) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.SetAuthor("Myrmex ERP", false)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
	pdf.Ln(4)

	return pdf
}

// pdfBytes finalises the PDF and returns its bytes.
func pdfBytes(pdf *gofpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// parseSemID converts a semester ID string to uuid.UUID.
// Empty string returns uuid.Nil (meaning "all semesters").
func parseSemID(semesterID string) (uuid.UUID, error) {
	if semesterID == "" {
		return uuid.Nil, nil
	}
	id, err := uuid.Parse(semesterID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid semester_id %q: %w", semesterID, err)
	}
	return id, nil
}
