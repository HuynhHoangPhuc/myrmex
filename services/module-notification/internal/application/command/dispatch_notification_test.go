package command_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
)

// stubPrefRepo implements repository.PreferenceRepository for testing.
type stubPrefRepo struct {
	enabled bool
	err     error
}

func (s *stubPrefRepo) GetByUser(_ context.Context, _ string) ([]entity.Preference, error) {
	return nil, nil
}

func (s *stubPrefRepo) BulkUpsert(_ context.Context, _ string, _ []entity.Preference) error {
	return nil
}

func (s *stubPrefRepo) IsEnabled(_ context.Context, _, _, _ string) (bool, error) {
	return s.enabled, s.err
}

// stubPushPublisher implements command.PushPublisher for testing.
type stubPushPublisher struct {
	calls int
}

func (s *stubPushPublisher) PublishPush(_, _, _, _, _ string, _ time.Time, _ int64) {
	s.calls++
}

func TestDispatch_SkipsWhenPreferenceDisabled(t *testing.T) {
	// nil pool is safe: Insert is never reached when preference is disabled.
	notifRepo := persistence.NewNotificationRepository(nil)
	prefRepo := &stubPrefRepo{enabled: false}
	publisher := &stubPushPublisher{}

	cmd := command.NewDispatchNotificationCommand(notifRepo, prefRepo, publisher, zap.NewNop())

	id, err := cmd.Execute(context.Background(), command.DispatchInput{
		UserID:  "user-1",
		Type:    entity.EventGradePosted,
		Channel: entity.ChannelInApp,
		Title:   "Grade Posted",
		Body:    "Your grade has been posted.",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "" {
		t.Errorf("expected empty id when preference disabled, got %q", id)
	}
	if publisher.calls != 0 {
		t.Errorf("expected no push published when skipped, got %d calls", publisher.calls)
	}
}

func TestDispatch_EmailChannel_SkipsWhenPreferenceDisabled(t *testing.T) {
	notifRepo := persistence.NewNotificationRepository(nil)
	prefRepo := &stubPrefRepo{enabled: false}
	publisher := &stubPushPublisher{}

	cmd := command.NewDispatchNotificationCommand(notifRepo, prefRepo, publisher, zap.NewNop())

	id, err := cmd.Execute(context.Background(), command.DispatchInput{
		UserID:  "user-2",
		Type:    entity.EventEnrollmentApproved,
		Channel: entity.ChannelEmail,
		Title:   "Enrollment Approved",
		Body:    "Your enrollment has been approved.",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "" {
		t.Errorf("expected empty id when preference disabled, got %q", id)
	}
	if publisher.calls != 0 {
		t.Errorf("expected no push published for email channel skip, got %d calls", publisher.calls)
	}
}
