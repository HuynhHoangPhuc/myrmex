package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type Claims struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id,omitempty"` // set for dept_head, teacher
	TeacherID    string `json:"teacher_id,omitempty"`    // set when user has linked teacher record
	jwt.RegisteredClaims
}

func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken creates an access token with role and optional dept/teacher scope.
// Pass empty strings for departmentID/teacherID when not applicable.
func (s *JWTService) GenerateAccessToken(userID, role, departmentID, teacherID string) (string, error) {
	claims := Claims{
		UserID:       userID,
		Role:         role,
		DepartmentID: departmentID,
		TeacherID:    teacherID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateInternalToken creates a long-lived service JWT for internal inter-component calls.
// Uses a fixed 24h expiry independent of the configured access token expiry.
func (s *JWTService) GenerateInternalToken() (string, error) {
	claims := Claims{
		UserID: "internal-service",
		Role:   "service",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// Validate implements the pkg/middleware TokenValidator interface.
func (s *JWTService) Validate(token string) (string, error) {
	claims, err := s.ValidateToken(token)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// ExtractClaims implements the pkg/middleware ClaimsExtractor interface.
// Returns userID, role, departmentID for richer gRPC context injection.
func (s *JWTService) ExtractClaims(token string) (userID, role, departmentID string, err error) {
	claims, err := s.ValidateToken(token)
	if err != nil {
		return "", "", "", err
	}
	return claims.UserID, claims.Role, claims.DepartmentID, nil
}
