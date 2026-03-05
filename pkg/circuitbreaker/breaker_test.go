package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/HuynhHoangPhuc/myrmex/pkg/circuitbreaker"
)

var errFake = errors.New("fake error")

func TestBreaker_ClosedByDefault(t *testing.T) {
	b := circuitbreaker.New(3, 100*time.Millisecond, 1)
	if b.State() != "closed" {
		t.Fatalf("expected closed, got %s", b.State())
	}
}

func TestBreaker_OpensAfterThreshold(t *testing.T) {
	b := circuitbreaker.New(3, 100*time.Millisecond, 1)
	fail := func() error { return errFake }

	for i := 0; i < 3; i++ {
		_ = b.Execute(fail)
	}

	if b.State() != "open" {
		t.Fatalf("expected open after threshold, got %s", b.State())
	}
}

func TestBreaker_FastFailsWhenOpen(t *testing.T) {
	b := circuitbreaker.New(2, 10*time.Second, 1)
	fail := func() error { return errFake }

	_ = b.Execute(fail)
	_ = b.Execute(fail)

	err := b.Execute(fail)
	if !errors.Is(err, circuitbreaker.ErrOpen) {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestBreaker_TransitionsHalfOpenAfterTimeout(t *testing.T) {
	b := circuitbreaker.New(2, 50*time.Millisecond, 1)
	fail := func() error { return errFake }

	_ = b.Execute(fail)
	_ = b.Execute(fail) // open

	time.Sleep(60 * time.Millisecond)

	// First call after timeout should be allowed (half-open probe)
	called := false
	_ = b.Execute(func() error {
		called = true
		return nil
	})

	if !called {
		t.Fatal("expected probe call to be executed in half-open state")
	}
}

func TestBreaker_ClosesAfterSuccessfulProbe(t *testing.T) {
	b := circuitbreaker.New(2, 50*time.Millisecond, 1)
	fail := func() error { return errFake }

	_ = b.Execute(fail)
	_ = b.Execute(fail) // open

	time.Sleep(60 * time.Millisecond)

	_ = b.Execute(func() error { return nil }) // successful probe

	if b.State() != "closed" {
		t.Fatalf("expected closed after successful probe, got %s", b.State())
	}
}

func TestBreaker_ReopensOnFailedProbe(t *testing.T) {
	b := circuitbreaker.New(2, 50*time.Millisecond, 1)
	fail := func() error { return errFake }

	_ = b.Execute(fail)
	_ = b.Execute(fail) // open

	time.Sleep(60 * time.Millisecond)

	_ = b.Execute(fail) // failed probe — should reopen

	if b.State() != "open" {
		t.Fatalf("expected open after failed probe, got %s", b.State())
	}
}

func TestBreaker_ResetsOnSuccess(t *testing.T) {
	b := circuitbreaker.New(3, 100*time.Millisecond, 1)
	fail := func() error { return errFake }
	succeed := func() error { return nil }

	_ = b.Execute(fail)
	_ = b.Execute(fail) // 2 failures (below threshold)
	_ = b.Execute(succeed)

	if b.State() != "closed" {
		t.Fatalf("expected closed after success resets failures, got %s", b.State())
	}

	// After reset, need threshold failures again to open
	_ = b.Execute(fail)
	_ = b.Execute(fail)
	if b.State() == "open" {
		t.Fatal("should not open yet — counter was reset by previous success")
	}
}
