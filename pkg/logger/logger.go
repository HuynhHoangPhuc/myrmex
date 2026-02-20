package logger

import "go.uber.org/zap"

// New creates a Zap logger based on environment.
// "production" uses JSON format; others use colored console.
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
