package messaging

import (
	"context"

	"github.com/HuynhHoangPhuc/myrmex/pkg/cache"
	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
)

// PrerequisiteCacheInvalidator clears cached prerequisite graphs after subject changes.
type PrerequisiteCacheInvalidator struct {
	cache cache.Cache
}

func NewPrerequisiteCacheInvalidator(c cache.Cache) *PrerequisiteCacheInvalidator {
	return &PrerequisiteCacheInvalidator{cache: c}
}

// Start subscribes to subject change events and invalidates the prerequisite cache.
// Subscriptions run in background goroutines managed by consumer until Close().
func (i *PrerequisiteCacheInvalidator) Start(ctx context.Context, consumer messaging.Consumer) error {
	subjects := []string{
		"subject.updated",
		"subject.deleted",
		"subject.prerequisite.added",
		"subject.prerequisite.removed",
	}
	for _, subject := range subjects {
		durable := "prereq-cache-invalidator-" + subject
		sub := subject // capture for closure
		if err := consumer.Subscribe(ctx, durable, sub, func(_ *messaging.Message) error {
			_ = i.cache.InvalidateByPattern(ctx, "prereq:subject:*")
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
