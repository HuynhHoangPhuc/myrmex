package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	"github.com/HuynhHoangPhuc/myrmex/pkg/cache"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// SubjectLookupClient reads subject details for prerequisite responses.
type SubjectLookupClient interface {
	GetSubject(ctx context.Context, in *subjectv1.GetSubjectRequest, opts ...grpc.CallOption) (*subjectv1.GetSubjectResponse, error)
}

// PrerequisiteLookupClient reads prerequisite edges from the subject module.
type PrerequisiteLookupClient interface {
	ListPrerequisites(ctx context.Context, in *subjectv1.ListPrerequisitesRequest, opts ...grpc.CallOption) (*subjectv1.ListPrerequisitesResponse, error)
}

// MissingPrerequisite describes a blocking prerequisite for enrollment.
type MissingPrerequisite struct {
	SubjectID   uuid.UUID
	SubjectCode string
	SubjectName string
	Type        string
}

type cachedPrerequisite struct {
	SubjectID string `json:"subject_id"`
	Type      string `json:"type"`
}

// PrerequisiteChecker validates whether a student can enroll in a subject.
type PrerequisiteChecker struct {
	enrollments    repository.EnrollmentRepository
	subjects       SubjectLookupClient
	prerequisites  PrerequisiteLookupClient
	cache          cache.Cache
	cacheTTL       time.Duration
}

func NewPrerequisiteChecker(
	enrollments repository.EnrollmentRepository,
	subjects SubjectLookupClient,
	prerequisites PrerequisiteLookupClient,
	cache cache.Cache,
	cacheTTL time.Duration,
) *PrerequisiteChecker {
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}
	return &PrerequisiteChecker{
		enrollments:   enrollments,
		subjects:      subjects,
		prerequisites: prerequisites,
		cache:         cache,
		cacheTTL:      cacheTTL,
	}
}

func (p *PrerequisiteChecker) Check(ctx context.Context, studentID, subjectID uuid.UUID) ([]MissingPrerequisite, error) {
	if p.enrollments == nil {
		return nil, fmt.Errorf("enrollment repository is required")
	}

	passedSubjectIDs, err := p.enrollments.ListPassedSubjectIDs(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("list passed subjects: %w", err)
	}

	prereqs, err := p.loadPrerequisites(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	passedSet := make(map[uuid.UUID]struct{}, len(passedSubjectIDs))
	for _, passedID := range passedSubjectIDs {
		passedSet[passedID] = struct{}{}
	}

	missing := make([]MissingPrerequisite, 0)
	for _, prereq := range prereqs {
		if prereq.Type != "hard" && prereq.Type != "strict" {
			continue
		}

		prereqID, err := uuid.Parse(prereq.SubjectID)
		if err != nil {
			return nil, fmt.Errorf("parse prerequisite subject_id %q: %w", prereq.SubjectID, err)
		}
		if _, ok := passedSet[prereqID]; ok {
			continue
		}

		missingPrereq := MissingPrerequisite{SubjectID: prereqID, Type: "strict"}
		if p.subjects != nil {
			resp, err := p.subjects.GetSubject(ctx, &subjectv1.GetSubjectRequest{Id: prereqID.String()})
			if err == nil && resp.GetSubject() != nil {
				missingPrereq.SubjectCode = resp.Subject.GetCode()
				missingPrereq.SubjectName = resp.Subject.GetName()
			}
		}
		missing = append(missing, missingPrereq)
	}

	return missing, nil
}

func (p *PrerequisiteChecker) loadPrerequisites(ctx context.Context, subjectID uuid.UUID) ([]cachedPrerequisite, error) {
	cacheKey := fmt.Sprintf("prereq:subject:%s", subjectID)
	if p.cache != nil {
		var cached []cachedPrerequisite
		err := p.cache.Get(ctx, cacheKey, &cached)
		switch {
		case err == nil:
			return cached, nil
		case errors.Is(err, cache.ErrCacheMiss):
			// Fetch from upstream below.
		default:
			return nil, fmt.Errorf("load prerequisite cache: %w", err)
		}
	}

	if p.prerequisites == nil {
		return nil, fmt.Errorf("prerequisite service unavailable")
	}

	resp, err := p.prerequisites.ListPrerequisites(ctx, &subjectv1.ListPrerequisitesRequest{SubjectId: subjectID.String()})
	if err != nil {
		return nil, fmt.Errorf("list prerequisites: %w", err)
	}

	mapped := make([]cachedPrerequisite, len(resp.Prerequisites))
	for i, prereq := range resp.Prerequisites {
		mapped[i] = cachedPrerequisite{
			SubjectID: prereq.GetPrerequisiteId(),
			Type:      prereq.GetType(),
		}
	}

	if p.cache != nil {
		_ = p.cache.Set(ctx, cacheKey, mapped, p.cacheTTL)
	}

	return mapped, nil
}
