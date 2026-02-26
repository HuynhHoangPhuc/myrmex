package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/application/export"
)

// ExportHandler serves analytics data as downloadable PDF or Excel files.
type ExportHandler struct {
	pdfGen   *export.PDFGenerator
	excelGen *export.ExcelGenerator
}

// NewExportHandler constructs an ExportHandler wiring PDF and Excel generators.
func NewExportHandler(pdfGen *export.PDFGenerator, excelGen *export.ExcelGenerator) *ExportHandler {
	return &ExportHandler{pdfGen: pdfGen, excelGen: excelGen}
}

// HandleExport reads query params and streams the generated file to the client.
//
// Query params:
//   - format     : "pdf" | "xlsx"  (required)
//   - type       : "workload" | "utilization" | "schedule"  (required)
//   - semester_id: UUID string (optional — omit for all semesters)
func (h *ExportHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	format := q.Get("format")
	reportType := q.Get("type")
	semesterID := q.Get("semester_id")

	if format != "pdf" && format != "xlsx" {
		http.Error(w, `query param "format" must be "pdf" or "xlsx"`, http.StatusBadRequest)
		return
	}
	if reportType != "workload" && reportType != "utilization" && reportType != "schedule" {
		http.Error(w, `query param "type" must be "workload", "utilization", or "schedule"`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	var (
		data        []byte
		err         error
		contentType string
		filename    string
	)

	switch format {
	case "pdf":
		data, err = h.dispatchPDF(ctx, reportType, semesterID)
		contentType = "application/pdf"
		filename = fmt.Sprintf("%s-report.pdf", reportType)
	default: // "xlsx"
		data, err = h.dispatchExcel(ctx, reportType, semesterID)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = fmt.Sprintf("%s-report.xlsx", reportType)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("generate report: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data) //nolint:errcheck
}

// dispatchPDF routes to the correct PDF generator method by report type.
func (h *ExportHandler) dispatchPDF(ctx context.Context, reportType, semesterID string) ([]byte, error) {
	switch reportType {
	case "workload":
		return h.pdfGen.GenerateWorkloadReport(ctx, semesterID)
	case "utilization":
		return h.pdfGen.GenerateUtilizationReport(ctx, semesterID)
	default: // "schedule"
		return h.pdfGen.GenerateScheduleReport(ctx, semesterID)
	}
}

// dispatchExcel routes to the correct Excel generator method by report type.
func (h *ExportHandler) dispatchExcel(ctx context.Context, reportType, semesterID string) ([]byte, error) {
	switch reportType {
	case "workload":
		return h.excelGen.GenerateWorkloadReport(ctx, semesterID)
	case "utilization":
		return h.excelGen.GenerateUtilizationReport(ctx, semesterID)
	default: // "schedule"
		return h.excelGen.GenerateScheduleReport(ctx, semesterID)
	}
}
