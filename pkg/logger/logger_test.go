package logger

import "testing"

func TestNew_ReturnsLogger(t *testing.T) {
	logger, err := New("development")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if syncErr := logger.Sync(); syncErr != nil {
		_ = syncErr
	}
}
