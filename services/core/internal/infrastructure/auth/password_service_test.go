package auth

import (
	"strings"
	"testing"
)

func TestPasswordService_HashAndVerify(t *testing.T) {
	svc := NewPasswordService()
	hash, err := svc.Hash("correct-horse-battery")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") && !strings.HasPrefix(hash, "$2y$") {
		prefix := hash
		if len(prefix) > 10 {
			prefix = prefix[:10]
		}
		t.Errorf("expected bcrypt hash prefix, got: %s", prefix)
	}
	if err := svc.Verify(hash, "correct-horse-battery"); err != nil {
		t.Errorf("verify correct password: %v", err)
	}
	if err := svc.Verify(hash, "wrong-password"); err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestPasswordService_DifferentHashes(t *testing.T) {
	svc := NewPasswordService()
	h1, err := svc.Hash("pass")
	if err != nil {
		t.Fatalf("hash 1: %v", err)
	}
	h2, err := svc.Hash("pass")
	if err != nil {
		t.Fatalf("hash 2: %v", err)
	}
	if h1 == h2 {
		t.Error("same password should produce different hashes (salted)")
	}
}
