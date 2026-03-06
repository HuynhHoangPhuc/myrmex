package messaging

import "go.uber.org/zap"

// nopLogger returns a no-op zap.Logger for use in tests.
func nopLogger() *zap.Logger {
	return zap.NewNop()
}
