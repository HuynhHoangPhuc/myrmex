package http

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/command"
)

// ImportHandler handles CSV bulk import of teachers and students.
type ImportHandler struct {
	registerHandler *command.RegisterUserHandler
	teachers        hrv1.TeacherServiceClient
	students        studentv1.StudentServiceClient
}

func NewImportHandler(
	registerHandler *command.RegisterUserHandler,
	teachers hrv1.TeacherServiceClient,
	students studentv1.StudentServiceClient,
) *ImportHandler {
	return &ImportHandler{
		registerHandler: registerHandler,
		teachers:        teachers,
		students:        students,
	}
}

// importResult is the JSON response for a bulk import operation.
type importResult struct {
	Total   int           `json:"total"`
	Created int           `json:"created"`
	Skipped int           `json:"skipped"`
	Errors  []importError `json:"errors"`
}

type importError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// randomPassword generates a secure 24-char random password (imported users log in via OAuth).
func randomPassword() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// ImportTeachers handles POST /api/admin/import/teachers (multipart/form-data, field "file").
// CSV columns: email,full_name,employee_code,department_id,specializations,phone,max_hours_per_week
func (h *ImportHandler) ImportTeachers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	// Skip header row
	if _, err := reader.Read(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty or invalid CSV"})
		return
	}

	ctx := c.Request.Context()
	result := importResult{}
	rowNum := 1 // data rows start at 2 (1 = header)

	for {
		rowNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: fmt.Sprintf("parse error: %v", err)})
			continue
		}
		if len(record) < 4 {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: "too few columns"})
			continue
		}
		result.Total++

		email, fullName, empCode, deptID := record[0], record[1], record[2], record[3]
		specializations := splitSemicolon(safeCol(record, 4))
		phone := safeCol(record, 5)
		maxHours, _ := strconv.Atoi(safeCol(record, 6))

		if email == "" || fullName == "" || deptID == "" {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: "email, full_name, department_id are required"})
			continue
		}
		if !strings.Contains(email, "@") {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: "invalid email"})
			continue
		}

		// Create core user — skip if already exists
		_, err = h.registerHandler.Handle(ctx, command.RegisterUserCommand{
			Email:    email,
			Password: randomPassword(),
			FullName: fullName,
			Role:     "teacher",
		})
		if err != nil {
			if isAlreadyExists(err) {
				result.Skipped++
				continue
			}
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: fmt.Sprintf("create user: %v", err)})
			continue
		}

		// Create teacher record in HR module
		_, err = h.teachers.CreateTeacher(ctx, &hrv1.CreateTeacherRequest{
			FullName:        fullName,
			Email:           email,
			DepartmentId:    deptID,
			EmployeeCode:    empCode,
			Specializations: specializations,
			Phone:           phone,
			MaxHoursPerWeek: int32(maxHours),
		})
		if err != nil {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: fmt.Sprintf("create teacher: %v", err)})
			continue
		}
		result.Created++
	}

	c.JSON(http.StatusOK, result)
}

// ImportStudents handles POST /api/admin/import/students (multipart/form-data, field "file").
// CSV columns: email,full_name,student_code,department_id,enrollment_year
func (h *ImportHandler) ImportStudents(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	if _, err := reader.Read(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty or invalid CSV"})
		return
	}

	ctx := c.Request.Context()
	result := importResult{}
	rowNum := 1

	for {
		rowNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: fmt.Sprintf("parse error: %v", err)})
			continue
		}
		if len(record) < 4 {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: "too few columns"})
			continue
		}
		result.Total++

		email, fullName, stuCode, deptID := record[0], record[1], record[2], record[3]
		year, _ := strconv.Atoi(safeCol(record, 4))

		if email == "" || fullName == "" || stuCode == "" || deptID == "" {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: "email, full_name, student_code, department_id are required"})
			continue
		}
		if !strings.Contains(email, "@") {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: "invalid email"})
			continue
		}

		_, err = h.registerHandler.Handle(ctx, command.RegisterUserCommand{
			Email:    email,
			Password: randomPassword(),
			FullName: fullName,
			Role:     "student",
		})
		if err != nil {
			if isAlreadyExists(err) {
				result.Skipped++
				continue
			}
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: fmt.Sprintf("create user: %v", err)})
			continue
		}

		_, err = h.students.CreateStudent(ctx, &studentv1.CreateStudentRequest{
			StudentCode:    stuCode,
			FullName:       fullName,
			Email:          email,
			DepartmentId:   deptID,
			EnrollmentYear: int32(year),
		})
		if err != nil {
			result.Errors = append(result.Errors, importError{Row: rowNum, Message: fmt.Sprintf("create student: %v", err)})
			continue
		}
		result.Created++
	}

	c.JSON(http.StatusOK, result)
}

// TeacherTemplate returns an empty CSV with teacher column headers.
func (h *ImportHandler) TeacherTemplate(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=teachers-template.csv")
	c.Header("Content-Type", "text/csv")
	c.String(http.StatusOK, "email,full_name,employee_code,department_id,specializations,phone,max_hours_per_week\n")
}

// StudentTemplate returns an empty CSV with student column headers.
func (h *ImportHandler) StudentTemplate(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=students-template.csv")
	c.Header("Content-Type", "text/csv")
	c.String(http.StatusOK, "email,full_name,student_code,department_id,enrollment_year\n")
}

// safeCol returns the column at index i or "" if out of range.
func safeCol(record []string, i int) string {
	if i < len(record) {
		return strings.TrimSpace(record[i])
	}
	return ""
}

// splitSemicolon splits a semicolon-separated string, ignoring empty parts.
func splitSemicolon(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ";")
	out := parts[:0]
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// isAlreadyExists returns true when the error indicates a duplicate email (gRPC AlreadyExists or DB unique violation).
func isAlreadyExists(err error) bool {
	if s, ok := status.FromError(err); ok {
		return s.Code() == codes.AlreadyExists
	}
	return strings.Contains(err.Error(), "already exists") ||
		strings.Contains(err.Error(), "unique") ||
		strings.Contains(err.Error(), "duplicate")
}
