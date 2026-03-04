package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	userIDKey     contextKey = "user_id"
	userRoleKey   contextKey = "user_role"
	departmentKey contextKey = "department_id"
)

// UserIDFromContext extracts the user ID set by AuthUnaryInterceptor.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

// UserRoleFromContext extracts the role set by AuthUnaryInterceptor.
func UserRoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userRoleKey).(string)
	return v, ok
}

// DepartmentIDFromContext extracts the department_id set by AuthUnaryInterceptor.
func DepartmentIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(departmentKey).(string)
	return v, ok
}

// TokenValidator validates a bearer token and returns the user ID.
type TokenValidator interface {
	Validate(token string) (userID string, err error)
}

// ClaimsExtractor extends TokenValidator to provide full claims (role, dept) from a token.
type ClaimsExtractor interface {
	TokenValidator
	ExtractClaims(token string) (userID, role, departmentID string, err error)
}

// AuthUnaryInterceptor returns a gRPC unary interceptor that validates bearer tokens
// and injects user_id, user_role, and department_id into the context.
func AuthUnaryInterceptor(validator TokenValidator, skipMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		// If the validator also supports claim extraction, use it for richer context
		if extractor, ok := validator.(ClaimsExtractor); ok {
			userID, role, deptID, err := extractor.ExtractClaims(tokens[0])
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			}
			ctx = context.WithValue(ctx, userIDKey, userID)
			ctx = context.WithValue(ctx, userRoleKey, role)
			ctx = context.WithValue(ctx, departmentKey, deptID)
			return handler(ctx, req)
		}

		// Fallback: only extract userID
		userID, err := validator.Validate(tokens[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		ctx = context.WithValue(ctx, userIDKey, userID)
		return handler(ctx, req)
	}
}
