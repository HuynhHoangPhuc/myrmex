package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestJWTService_AccessToken_RoundTrip(t *testing.T) {
	svc := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	tok, err := svc.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID: got %q want %q", claims.UserID, "user-1")
	}
	if claims.Role != "admin" {
		t.Errorf("Role: got %q want %q", claims.Role, "admin")
	}
}

func TestJWTService_RefreshToken_NoRole(t *testing.T) {
	svc := NewJWTService("secret", time.Hour, 7*24*time.Hour)
	tok, err := svc.GenerateRefreshToken("user-2")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != "user-2" {
		t.Errorf("UserID: got %q want %q", claims.UserID, "user-2")
	}
	if claims.Role != "" {
		t.Errorf("refresh token should have empty Role, got %q", claims.Role)
	}
}

func TestJWTService_Expired(t *testing.T) {
	svc := NewJWTService("secret", -1*time.Second, time.Hour)
	tok, err := svc.GenerateAccessToken("u", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	_, err = svc.ValidateToken(tok)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestJWTService_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-a", time.Hour, time.Hour)
	svc2 := NewJWTService("secret-b", time.Hour, time.Hour)
	tok, err := svc1.GenerateAccessToken("u", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	_, err = svc2.ValidateToken(tok)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestJWTService_TamperedPayload(t *testing.T) {
	svc := NewJWTService("secret", time.Hour, time.Hour)
	tok, err := svc.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatal("malformed token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	modified := strings.Replace(string(payload), "admin", "superadmin", 1)
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(modified))
	tampered := strings.Join(parts, ".")
	_, err = svc.ValidateToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestJWTService_AlgorithmConfusion(t *testing.T) {
	svc := NewJWTService("secret", time.Hour, time.Hour)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"user_id":"x"}`))
	fakeToken := header + "." + payload + ".fakesig"
	_, err := svc.ValidateToken(fakeToken)
	if err == nil {
		t.Fatal("expected error for RS256 algorithm")
	}
}

func TestJWTService_Validate_Wrapper(t *testing.T) {
	svc := NewJWTService("secret", time.Hour, time.Hour)
	tok, err := svc.GenerateAccessToken("user-42", "viewer")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	userID, err := svc.Validate(tok)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if userID != "user-42" {
		t.Errorf("got %q want %q", userID, "user-42")
	}
}
