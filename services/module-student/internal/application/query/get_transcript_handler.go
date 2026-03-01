package query

import (
	"context"
	"fmt"
	"math"
	"time"

	corev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/core/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	"github.com/HuynhHoangPhuc/myrmex/pkg/cache"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const transcriptCacheTTL = 5 * time.Minute

// SubjectListClient reads subject metadata used for transcript enrichment.
type SubjectListClient interface {
	ListSubjects(ctx context.Context, in *subjectv1.ListSubjectsRequest, opts ...grpc.CallOption) (*subjectv1.ListSubjectsResponse, error)
}

// GetStudentTranscriptQuery carries the student ID to load.
type GetStudentTranscriptQuery struct {
	StudentID uuid.UUID
}

// GetStudentTranscriptHandler assembles a student's transcript.
type GetStudentTranscriptHandler struct {
	students repository.StudentRepository
	grades   repository.GradeRepository
	subjects SubjectListClient
	cache    cache.Cache
}

func NewGetStudentTranscriptHandler(
	students repository.StudentRepository,
	grades repository.GradeRepository,
	subjects SubjectListClient,
	cache cache.Cache,
) *GetStudentTranscriptHandler {
	return &GetStudentTranscriptHandler{students: students, grades: grades, subjects: subjects, cache: cache}
}

func (h *GetStudentTranscriptHandler) Handle(ctx context.Context, q GetStudentTranscriptQuery) (*entity.StudentTranscript, error) {
	if h.students == nil || h.grades == nil {
		return nil, fmt.Errorf("repositories are required")
	}

	if h.cache != nil {
		var cached entity.StudentTranscript
		if err := h.cache.Get(ctx, transcriptCacheKey(q.StudentID), &cached); err == nil {
			return &cached, nil
		}
	}

	student, err := h.students.GetByID(ctx, q.StudentID)
	if err != nil {
		return nil, fmt.Errorf("get student: %w", err)
	}
	records, err := h.grades.GetStudentTranscript(ctx, q.StudentID)
	if err != nil {
		return nil, fmt.Errorf("get transcript rows: %w", err)
	}

	subjectsByID := make(map[string]*subjectv1.Subject)
	if h.subjects != nil {
		resp, err := h.subjects.ListSubjects(ctx, &subjectv1.ListSubjectsRequest{Pagination: &corev1.PaginationRequest{Page: 1, PageSize: 1000}})
		if err == nil {
			for _, subject := range resp.GetSubjects() {
				subjectsByID[subject.GetId()] = subject
			}
		}
	}

	entries := make([]entity.TranscriptEntry, len(records))
	var totalWeight float64
	var totalCredits int32
	var passedCredits int32
	for i, record := range records {
		entry := entity.TranscriptEntry{
			EnrollmentID: record.EnrollmentID.String(),
			SemesterID:   record.SemesterID.String(),
			SubjectID:    record.SubjectID.String(),
			Status:       record.Status,
			GradedAt:     record.GradedAt,
		}
		if record.GradeNumeric != nil {
			entry.GradeNumeric = *record.GradeNumeric
		}
		if record.GradeLetter != nil {
			entry.GradeLetter = *record.GradeLetter
		}
		if subject := subjectsByID[entry.SubjectID]; subject != nil {
			entry.SubjectCode = subject.GetCode()
			entry.SubjectName = subject.GetName()
			entry.Credits = subject.GetCredits()
		}
		entries[i] = entry
		totalCredits += entry.Credits
		if entry.GradeNumeric > 0 {
			totalWeight += entry.GradeNumeric * float64(entry.Credits)
			if entry.GradeNumeric >= 4.0 {
				passedCredits += entry.Credits
			}
		}
	}

	gpa := 0.0
	if totalCredits > 0 {
		gpa = math.Round((totalWeight/float64(totalCredits))*100) / 100
	}

	transcript := &entity.StudentTranscript{
		Student:       student,
		Entries:       entries,
		GPA:           gpa,
		TotalCredits:  totalCredits,
		PassedCredits: passedCredits,
	}
	if h.cache != nil {
		_ = h.cache.Set(ctx, transcriptCacheKey(q.StudentID), transcript, transcriptCacheTTL)
	}
	return transcript, nil
}

func transcriptCacheKey(studentID uuid.UUID) string {
	return fmt.Sprintf("student:transcript:%s", studentID)
}
