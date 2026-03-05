// Package circuitbreaker provides a thread-safe circuit breaker for wrapping
// calls to external dependencies (Redis, Pub/Sub, SMTP).
// States: Closed (normal) → Open (fast-fail after N failures) → Half-Open (probe) → Closed
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit is open and calls are being fast-failed.
var ErrOpen = errors.New("circuit breaker open")

type state int

const (
	stateClosed   state = iota // normal operation
	stateOpen                  // fast-fail; waiting for timeout
	stateHalfOpen              // allow one probe request
)

// Breaker is a thread-safe circuit breaker.
type Breaker struct {
	mu           sync.Mutex
	state        state
	failures     int
	threshold    int           // consecutive failures before opening
	timeout      time.Duration // duration to wait before half-open probe
	openedAt     time.Time
	maxHalfOpen  int // max concurrent probes in half-open state
	halfOpenCount int
}

// New creates a Breaker with the given configuration.
//   - threshold: consecutive failures before opening the circuit
//   - timeout: how long to wait before allowing a probe (half-open)
//   - maxHalfOpen: max concurrent probe calls in half-open state
func New(threshold int, timeout time.Duration, maxHalfOpen int) *Breaker {
	return &Breaker{
		threshold:   threshold,
		timeout:     timeout,
		maxHalfOpen: maxHalfOpen,
	}
}

// Execute runs fn within the circuit breaker.
// Returns ErrOpen immediately if the circuit is open and the probe window has not elapsed.
func (b *Breaker) Execute(fn func() error) error {
	if err := b.before(); err != nil {
		return err
	}
	err := fn()
	b.after(err)
	return err
}

// before checks if the call is allowed and transitions state as needed.
func (b *Breaker) before() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case stateClosed:
		return nil

	case stateOpen:
		if time.Since(b.openedAt) < b.timeout {
			return ErrOpen
		}
		// Timeout elapsed — transition to half-open and allow one probe
		b.state = stateHalfOpen
		b.halfOpenCount = 0
		fallthrough

	case stateHalfOpen:
		if b.halfOpenCount >= b.maxHalfOpen {
			return ErrOpen
		}
		b.halfOpenCount++
		return nil
	}
	return nil
}

// after records success or failure and transitions state accordingly.
func (b *Breaker) after(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err == nil {
		// Success — reset to closed
		b.failures = 0
		b.state = stateClosed
		b.halfOpenCount = 0
		return
	}

	b.failures++
	if b.state == stateHalfOpen || b.failures >= b.threshold {
		// Open the circuit
		b.state = stateOpen
		b.openedAt = time.Now()
		b.failures = 0
		b.halfOpenCount = 0
	}
}

// State returns the current state label for logging/metrics.
func (b *Breaker) State() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case stateOpen:
		return "open"
	case stateHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}
