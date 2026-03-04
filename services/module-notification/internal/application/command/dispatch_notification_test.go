package command_test

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
)

// stubPrefRepo implements repository.PreferenceRepository for testing.
type stubPrefRepo struct {
	enabled bool
}

func (s *stubPrefRepo) GetByUser(_ context.Context, _ string) ([]entity.Preference, error) {
	return nil, nil
}

func (s *stubPrefRepo) BulkUpsert(_ context.Context, _ string, _ []entity.Preference) error {
	return nil
}

func (s *stubPrefRepo) IsEnabled(_ context.Context, _, _, _ string) (bool, error) {
	return s.enabled, nil
}

func TestDispatch_SkipsWhenPreferenceDisabled(t *testing.T) {
	// nil pool is safe: Insert is never reached when preference is disabled
	notifRepo := persistence.NewNotificationRepository(nil)
	prefRepo := &stubPrefRepo{enabled: false}
	publisher := messaging.NewNATSPublisher(nil, zap.NewNop())

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
}
