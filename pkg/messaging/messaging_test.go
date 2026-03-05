package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewMessage tests the NewMessage constructor.
func TestNewMessage(t *testing.T) {
	ackCalled := false
	nackCalled := false

	ack := func() error {
		ackCalled = true
		return nil
	}
	nack := func() error {
		nackCalled = true
		return nil
	}

	msg := NewMessage("test.subject", []byte("test data"), ack, nack)

	assert.NotNil(t, msg)
	assert.Equal(t, "test.subject", msg.Subject)
	assert.Equal(t, []byte("test data"), msg.Data)

	// Test that ack and nack callbacks are set
	err := msg.Ack()
	assert.NoError(t, err)
	assert.True(t, ackCalled)

	err = msg.Nack()
	assert.NoError(t, err)
	assert.True(t, nackCalled)
}

// TestMessage_AckNilCallback tests Ack() when callback is nil.
func TestMessage_AckNilCallback(t *testing.T) {
	msg := NewMessage("test.subject", []byte("test data"), nil, nil)

	err := msg.Ack()
	assert.NoError(t, err)
}

// TestMessage_NackNilCallback tests Nack() when callback is nil.
func TestMessage_NackNilCallback(t *testing.T) {
	msg := NewMessage("test.subject", []byte("test data"), nil, nil)

	err := msg.Nack()
	assert.NoError(t, err)
}

// TestMessage_AckError tests Ack() when callback returns error.
func TestMessage_AckError(t *testing.T) {
	expectedErr := errors.New("ack failed")
	ack := func() error {
		return expectedErr
	}

	msg := NewMessage("test.subject", []byte("test data"), ack, nil)

	err := msg.Ack()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// TestMessage_NackError tests Nack() when callback returns error.
func TestMessage_NackError(t *testing.T) {
	expectedErr := errors.New("nack failed")
	nack := func() error {
		return expectedErr
	}

	msg := NewMessage("test.subject", []byte("test data"), nil, nack)

	err := msg.Nack()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// TestNoopPublisher_Publish tests NoopPublisher.Publish returns nil.
func TestNoopPublisher_Publish(t *testing.T) {
	noop := &NoopPublisher{}
	ctx := context.Background()

	err := noop.Publish(ctx, "test.subject", []byte("test data"))
	assert.NoError(t, err)
}

// TestNoopPublisher_PublishIgnoresInput tests NoopPublisher ignores all inputs.
func TestNoopPublisher_PublishIgnoresInput(t *testing.T) {
	noop := &NoopPublisher{}
	ctx := context.Background()

	// Test with various inputs
	tests := []struct {
		subject string
		data    []byte
	}{
		{"", []byte{}},
		{"test.subject", []byte("data")},
		{"very.long.subject.with.many.parts", []byte("large data blob")},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			err := noop.Publish(ctx, tt.subject, tt.data)
			assert.NoError(t, err)
		})
	}
}

// TestNoopPublisher_Close tests NoopPublisher.Close returns nil.
func TestNoopPublisher_Close(t *testing.T) {
	noop := &NoopPublisher{}

	err := noop.Close()
	assert.NoError(t, err)
}

// TestNoopPublisher_ConcurrentCalls tests NoopPublisher with concurrent calls.
func TestNoopPublisher_ConcurrentCalls(t *testing.T) {
	noop := &NoopPublisher{}
	ctx := context.Background()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			err := noop.Publish(ctx, "test.subject", []byte("data"))
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
