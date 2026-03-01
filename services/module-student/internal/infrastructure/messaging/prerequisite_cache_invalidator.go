package messaging

import (
	"context"
	"fmt"

	"github.com/HuynhHoangPhuc/myrmex/pkg/cache"
	"github.com/nats-io/nats.go"
)

// PrerequisiteCacheInvalidator clears cached prerequisite graphs after subject changes.
type PrerequisiteCacheInvalidator struct {
	cache cache.Cache
}

func NewPrerequisiteCacheInvalidator(cache cache.Cache) *PrerequisiteCacheInvalidator {
	return &PrerequisiteCacheInvalidator{cache: cache}
}

func (i *PrerequisiteCacheInvalidator) Start(ctx context.Context, bus *NATSPublisher) ([]*nats.Subscription, error) {
	if i == nil || i.cache == nil || bus == nil {
		return nil, nil
	}

	handler := func(msg *nats.Msg) {
		_ = i.cache.InvalidateByPattern(ctx, "prereq:subject:*")
	}

	subjects := []string{
		"subject.updated",
		"subject.deleted",
		"subject.prerequisite.added",
		"subject.prerequisite.removed",
	}
	subscriptions := make([]*nats.Subscription, 0, len(subjects))
	for _, subject := range subjects {
		sub, err := bus.Subscribe(subject, handler)
		if err != nil {
			for _, existing := range subscriptions {
				_ = existing.Unsubscribe()
			}
			return nil, fmt.Errorf("subscribe prerequisite invalidator: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}
