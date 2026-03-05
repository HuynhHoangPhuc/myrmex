package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPushPublisher tests NewPushPublisher constructor.
func TestNewPushPublisher(t *testing.T) {
	// Test with nil redis client
	pp := NewPushPublisher(nil)
	assert.NotNil(t, pp)
	assert.Nil(t, pp.rdb)

	// Test with non-nil redis client would require actual redis instance
	// so we just test the constructor doesn't panic
}

// TestPushPublisher_PublishWithNilRedis tests Publish returns nil when rdb is nil.
func TestPushPublisher_PublishWithNilRedis(t *testing.T) {
	pp := NewPushPublisher(nil)
	ctx := context.Background()

	err := pp.Publish(ctx, "notification.push.user-123", []byte("test message"))
	assert.NoError(t, err)
}

// TestPushPublisher_PublishIgnoresInputWithNilRedis tests Publish ignores all inputs when rdb is nil.
func TestPushPublisher_PublishIgnoresInputWithNilRedis(t *testing.T) {
	pp := NewPushPublisher(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		channel string
		payload []byte
	}{
		{"empty channel", "", []byte{}},
		{"empty payload", "notification.push.user-123", []byte{}},
		{"normal message", "notification.push.user-123", []byte("test message")},
		{"large payload", "notification.push.user-123", make([]byte, 1024*1024)}, // 1MB
		{"special chars", "notification.push.user-123!@#$%", []byte("special chars")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pp.Publish(ctx, tt.channel, tt.payload)
			assert.NoError(t, err)
		})
	}
}

// TestPushPublisher_PublishWithCancelledContext tests Publish with cancelled context.
func TestPushPublisher_PublishWithCancelledContext(t *testing.T) {
	pp := NewPushPublisher(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should still return nil when rdb is nil, even with cancelled context
	err := pp.Publish(ctx, "notification.push.user-123", []byte("test"))
	assert.NoError(t, err)
}

// TestNewPushSubscriber tests NewPushSubscriber constructor.
func TestNewPushSubscriber(t *testing.T) {
	// Test with nil redis client
	ps := NewPushSubscriber(nil)
	assert.NotNil(t, ps)
	assert.Nil(t, ps.rdb)
}

// TestPushSubscriber_PSubscribeWithNilRedis tests PSubscribe returns error when rdb is nil.
func TestPushSubscriber_PSubscribeWithNilRedis(t *testing.T) {
	ps := NewPushSubscriber(nil)
	ctx := context.Background()

	ch, err := ps.PSubscribe(ctx, "notification.push.*")
	assert.Error(t, err)
	assert.Nil(t, ch)
	assert.Equal(t, "redis not available", err.Error())
}

// TestPushSubscriber_PSubscribeErrorMessage tests error message from PSubscribe.
func TestPushSubscriber_PSubscribeErrorMessage(t *testing.T) {
	ps := NewPushSubscriber(nil)
	ctx := context.Background()

	_, err := ps.PSubscribe(ctx, "any.pattern")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis not available")
}

// TestPushSubscriber_PSubscribeWithDifferentPatterns tests PSubscribe with various patterns.
func TestPushSubscriber_PSubscribeWithDifferentPatterns(t *testing.T) {
	ps := NewPushSubscriber(nil)
	ctx := context.Background()

	patterns := []string{
		"notification.push.*",
		"notification.*",
		"*",
		"notification.push.user-*",
		"notification.push.user-[0-9]*",
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			ch, err := ps.PSubscribe(ctx, pattern)
			assert.Error(t, err)
			assert.Nil(t, ch)
		})
	}
}

// TestPushMessage tests PushMessage struct.
func TestPushMessage(t *testing.T) {
	msg := PushMessage{
		Channel: "notification.push.user-123",
		Payload: []byte("test payload"),
	}

	assert.Equal(t, "notification.push.user-123", msg.Channel)
	assert.Equal(t, []byte("test payload"), msg.Payload)
}

// TestPushPublisher_ConcurrentPublish tests concurrent Publish calls with nil redis.
func TestPushPublisher_ConcurrentPublish(t *testing.T) {
	pp := NewPushPublisher(nil)
	ctx := context.Background()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			channel := "notification.push.user-" + string(rune(id))
			payload := []byte("message " + string(rune(id)))
			err := pp.Publish(ctx, channel, payload)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestPushSubscriber_ConcurrentPSubscribe tests concurrent PSubscribe calls with nil redis.
func TestPushSubscriber_ConcurrentPSubscribe(t *testing.T) {
	ps := NewPushSubscriber(nil)
	ctx := context.Background()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := ps.PSubscribe(ctx, "notification.push.*")
			assert.Error(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestPushPublisher_PublishWithContext tests Publish with various context scenarios.
func TestPushPublisher_PublishWithContext(t *testing.T) {
	pp := NewPushPublisher(nil)

	tests := []struct {
		name    string
		ctxFunc func() (context.Context, func())
	}{
		{
			name: "background context",
			ctxFunc: func() (context.Context, func()) {
				return context.Background(), func() {}
			},
		},
		{
			name: "context with timeout",
			ctxFunc: func() (context.Context, func()) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
		},
		{
			name: "context with deadline",
			ctxFunc: func() (context.Context, func()) {
				return context.WithDeadline(context.Background(), time.Now().Add(1*time.Second))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.ctxFunc()
			defer cancel()

			err := pp.Publish(ctx, "notification.push.user-123", []byte("test"))
			assert.NoError(t, err)
		})
	}
}

// TestPushPublisher_PublishReturnsBefore test shows immediate return with nil rdb.
func TestPushPublisher_PublishReturnsBefore(t *testing.T) {
	pp := NewPushPublisher(nil)
	ctx := context.Background()

	start := time.Now()
	err := pp.Publish(ctx, "notification.push.user-123", []byte("test"))
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Publish should return immediately")
}

// TestPushSubscriber_NilRedisCheck tests that PSubscribe correctly checks for nil rdb.
func TestPushSubscriber_NilRedisCheck(t *testing.T) {
	// Create subscriber with nil rdb
	ps := NewPushSubscriber(nil)
	ctx := context.Background()

	// Any pattern should fail with same error
	patterns := []string{"a", "b", "*", "notification.*"}
	for _, pattern := range patterns {
		_, err := ps.PSubscribe(ctx, pattern)
		assert.Error(t, err)
		assert.NotNil(t, err)
	}
}

// TestPushPublisher_ZeroValueRedis tests PushPublisher with zero-value struct.
func TestPushPublisher_ZeroValueRedis(t *testing.T) {
	// Create zero-value PushPublisher
	var pp PushPublisher
	ctx := context.Background()

	// Should handle nil rdb gracefully
	err := pp.Publish(ctx, "notification.push.user-123", []byte("test"))
	assert.NoError(t, err)
}

// BenchmarkPushPublisher_PublishWithNilRedis benchmarks Publish with nil rdb.
func BenchmarkPushPublisher_PublishWithNilRedis(b *testing.B) {
	pp := NewPushPublisher(nil)
	ctx := context.Background()
	payload := []byte("test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pp.Publish(ctx, "notification.push.user-123", payload)
	}
}

// BenchmarkPushSubscriber_PSubscribeWithNilRedis benchmarks PSubscribe with nil rdb.
func BenchmarkPushSubscriber_PSubscribeWithNilRedis(b *testing.B) {
	ps := NewPushSubscriber(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ps.PSubscribe(ctx, "notification.push.*")
	}
}
