package export

import (
	"bytes"
	"context"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// ExcelGenerator creates Excel reports from analytics data.
type ExcelGenerator struct {
	repo *persistence.AnalyticsRepository
}

// NewExcelGenerator constructs an ExcelGenerator with the given repository.
func NewExcelGenerator(repo *persistence.AnalyticsRepository) *ExcelGenerator {
	return &ExcelGenerator{repo: repo}
}

// GenerateWorkloadReport returns an Excel workbook with teacher workload data.
// Columns: Teacher | Department ID | Hours/Week | Total Hours
func (g *ExcelGenerator) GenerateWorkloadReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	stats, err := g.repo.GetWorkloadStats(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch workload stats: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck

	sheet := "Workload"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Teacher", "Department ID", "Hours/Week", "Total Hours"}
	if err := writeExcelHeader(f, sheet, headers); err != nil {
		return nil, err
	}

	for i, s := range stats {
		row := i + 2
		values := []any{
			s.TeacherName,
			s.DepartmentID.String(),
			s.HoursPerWeek,
			s.TotalHours,
		}
		if err := writeExcelRow(f, sheet, row, values); err != nil {
			return nil, err
		}
	}

	setColumnWidths(f, sheet, []float64{30, 38, 14, 14})

	return excelBytes(f)
}

// GenerateUtilizationReport returns an Excel workbook with department utilization data.
// Columns: Department | Assigned Slots | Total Slots | Utilization %
func (g *ExcelGenerator) GenerateUtilizationReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	stats, err := g.repo.GetUtilizationStats(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch utilization stats: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck

	sheet := "Utilization"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Department", "Assigned Slots", "Total Slots", "Utilization %"}
	if err := writeExcelHeader(f, sheet, headers); err != nil {
		return nil, err
	}

	for i, s := range stats {
		row := i + 2
		values := []any{
			s.DepartmentName,
			s.AssignedSlots,
			s.TotalSlots,
			s.UtilizationPct,
		}
		if err := writeExcelRow(f, sheet, row, values); err != nil {
			return nil, err
		}
	}

	setColumnWidths(f, sheet, []float64{30, 16, 14, 16})

	return excelBytes(f)
}

// GenerateScheduleReport returns an Excel workbook with schedule metrics data.
// Columns: Semester | Assigned Slots | Total Slots | Fill Rate %
func (g *ExcelGenerator) GenerateScheduleReport(ctx context.Context, semesterID string) ([]byte, error) {
	sid, err := parseSemID(semesterID)
	if err != nil {
		return nil, err
	}

	metrics, err := g.repo.GetScheduleMetrics(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("fetch schedule metrics: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck

	sheet := "Schedule"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Semester", "Assigned Slots", "Total Slots", "Fill Rate %"}
	if err := writeExcelHeader(f, sheet, headers); err != nil {
		return nil, err
	}

	for i, m := range metrics {
		fillRate := 0.0
		if m.TotalSlots > 0 {
			fillRate = float64(m.AssignedSlots) / float64(m.TotalSlots) * 100
		}
		row := i + 2
		values := []any{
			m.SemesterName,
			m.AssignedSlots,
			m.TotalSlots,
			fillRate,
		}
		if err := writeExcelRow(f, sheet, row, values); err != nil {
			return nil, err
		}
	}

	setColumnWidths(f, sheet, []float64{30, 16, 14, 14})

	return excelBytes(f)
}

// writeExcelHeader writes bold header row to row 1 of the sheet.
func writeExcelHeader(f *excelize.File, sheet string, headers []string) error {
	boldStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4287F5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		return fmt.Errorf("create header style: %w", err)
	}

	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return fmt.Errorf("set header cell %s: %w", cell, err)
		}
		if err := f.SetCellStyle(sheet, cell, cell, boldStyle); err != nil {
			return fmt.Errorf("set header style %s: %w", cell, err)
		}
	}
	return nil
}

// writeExcelRow writes a data row at the given 1-based row index.
func writeExcelRow(f *excelize.File, sheet string, row int, values []any) error {
	for col, v := range values {
		cell, _ := excelize.CoordinatesToCellName(col+1, row)
		if err := f.SetCellValue(sheet, cell, v); err != nil {
			return fmt.Errorf("set cell %s: %w", cell, err)
		}
	}
	return nil
}

// setColumnWidths sets character widths for columns A, B, C, ... in order.
func setColumnWidths(f *excelize.File, sheet string, widths []float64) {
	cols := []string{"A", "B", "C", "D", "E", "F"}
	for i, w := range widths {
		if i >= len(cols) {
			break
		}
		f.SetColWidth(sheet, cols[i], cols[i], w) //nolint:errcheck
	}
}

// excelBytes writes the workbook to a byte slice.
func excelBytes(f *excelize.File) ([]byte, error) {
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}
