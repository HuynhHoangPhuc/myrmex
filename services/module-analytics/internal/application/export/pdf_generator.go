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
// Table columns: Teacher | Department | Hours/Week | Total Hours
func (g *PDFGenerator) GenerateWorkloadReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	stats, err := g.repo.GetWorkloadStats(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch workload stats: %w", err)
	}

	pdf := newPDF("Teacher Workload Report")
	headers := []string{"Teacher", "Department ID", "Hours/Week", "Total Hours"}
	widths := []float64{60, 60, 35, 35}

	addTableHeader(pdf, headers, widths)

	for _, s := range stats {
		row := []string{
			s.TeacherName,
			s.DepartmentID.String(),
			fmt.Sprintf("%.1f", s.HoursPerWeek),
			fmt.Sprintf("%.1f", s.TotalHours),
		}
		addTableRow(pdf, row, widths)
	}

	return pdfBytes(pdf)
}

// GenerateUtilizationReport returns a PDF with department utilization data for the given semester.
// Table columns: Department | Assigned Slots | Total Slots | Utilization %
func (g *PDFGenerator) GenerateUtilizationReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	stats, err := g.repo.GetUtilizationStats(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch utilization stats: %w", err)
	}

	pdf := newPDF("Department Utilization Report")
	headers := []string{"Department", "Assigned Slots", "Total Slots", "Utilization %"}
	widths := []float64{70, 40, 35, 45}

	addTableHeader(pdf, headers, widths)

	for _, s := range stats {
		row := []string{
			s.DepartmentName,
			fmt.Sprintf("%d", s.AssignedSlots),
			fmt.Sprintf("%d", s.TotalSlots),
			fmt.Sprintf("%.1f%%", s.UtilizationPct),
		}
		addTableRow(pdf, row, widths)
	}

	return pdfBytes(pdf)
}

// GenerateScheduleReport returns a PDF with schedule metrics for the given semester.
// Table columns: Semester | Assigned Slots | Total Slots | Fill Rate %
func (g *PDFGenerator) GenerateScheduleReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	metrics, err := g.repo.GetScheduleMetrics(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch schedule metrics: %w", err)
	}

	pdf := newPDF("Schedule Metrics Report")
	headers := []string{"Semester", "Assigned Slots", "Total Slots", "Fill Rate %"}
	widths := []float64{70, 40, 35, 45}

	addTableHeader(pdf, headers, widths)

	for _, m := range metrics {
		fillRate := 0.0
		if m.TotalSlots > 0 {
			fillRate = float64(m.AssignedSlots) / float64(m.TotalSlots) * 100
		}
		row := []string{
			m.SemesterName,
			fmt.Sprintf("%d", m.AssignedSlots),
			fmt.Sprintf("%d", m.TotalSlots),
			fmt.Sprintf("%.1f%%", fillRate),
		}
		addTableRow(pdf, row, widths)
	}

	return pdfBytes(pdf)
}

// newPDF creates a new PDF document with standard layout and title.
func newPDF(title string) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.SetAuthor("Myrmex ERP", false)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
	pdf.Ln(4)

	return pdf
}

// addTableHeader writes a bold header row to the PDF table.
func addTableHeader(pdf *gofpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(66, 135, 245)
	pdf.SetTextColor(255, 255, 255)

	for i, h := range headers {
		pdf.CellFormat(widths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Reset text color for rows
	pdf.SetTextColor(0, 0, 0)
}

// addTableRow writes a single data row to the PDF table with alternating row color.
func addTableRow(pdf *gofpdf.Fpdf, row []string, widths []float64) {
	pdf.SetFont("Arial", "", 9)

	// Alternate row background
	if pdf.GetY() > 0 {
		pdf.SetFillColor(245, 245, 245)
	}

	for i, cell := range row {
		pdf.CellFormat(widths[i], 7, cell, "1", 0, "L", false, 0, "")
	}
	pdf.Ln(-1)
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
