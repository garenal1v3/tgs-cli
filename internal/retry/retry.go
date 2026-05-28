// Package retry provides generic retry logic with FLOOD_WAIT handling
// and exponential backoff for Telegram RPC calls.
package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"time"

	"github.com/gotd/td/tgerr"
)

// Policy configures retry behaviour.
type Policy struct {
	MaxRetries   int
	MaxFloodWait time.Duration
	BaseDelay    time.Duration
	MaxDelay     time.Duration
}

// DefaultPolicy returns a Policy with sensible defaults.
func DefaultPolicy() *Policy {
	return &Policy{
		MaxRetries:   3,
		MaxFloodWait: 60 * time.Second,
		BaseDelay:    500 * time.Millisecond,
		MaxDelay:     10 * time.Second,
	}
}

// FloodWaitError indicates Telegram asked us to wait before retrying.
type FloodWaitError struct {
	Seconds int
}

func (e *FloodWaitError) Error() string {
	return fmt.Sprintf("FLOOD_WAIT_%d", e.Seconds)
}

// PermanentError wraps an error that must NOT be retried.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// permanentTypes are RPC error types that should never be retried.
var permanentTypes = []string{
	"CHANNEL_PRIVATE",
	"CHAT_ADMIN_REQUIRED",
	"CHAT_ID_INVALID",
	"FOLDER_ID_INVALID",
	"MSG_ID_INVALID",
	"PEER_ID_INVALID",
	"USERNAME_INVALID",
	"USERNAME_NOT_OCCUPIED",
	"PHONE_NOT_OCCUPIED",
	"INPUT_FILTER_INVALID",
	"SEARCH_QUERY_EMPTY",
	"CHANNEL_INVALID",
}

// ClassifyError converts a gotd/td RPC error into a FloodWaitError,
// PermanentError, or returns it unchanged for transient errors.
func ClassifyError(err error) error {
	if err == nil {
		return nil
	}

	// Check for FLOOD_WAIT first.
	if d, ok := tgerr.AsFloodWait(err); ok {
		seconds := int(math.Ceil(d.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		return &FloodWaitError{Seconds: seconds}
	}

	// Check for known permanent error types.
	if tgerr.Is(err, permanentTypes...) {
		return &PermanentError{Err: err}
	}

	return err
}

// sleeper is an interface for sleeping, used to enable testing.
type sleeper interface {
	Sleep(ctx context.Context, d time.Duration) error
}

// realSleeper uses time.After for real sleeping.
type realSleeper struct{}

func (realSleeper) Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// defaultSleeper is the production sleeper.
var defaultSleeper sleeper = realSleeper{}

// Do retries fn according to the given Policy.
//
//   - On success, returns the result immediately.
//   - On PermanentError, returns immediately without retry.
//   - On FloodWaitError with seconds <= MaxFloodWait, waits and retries.
//   - On other errors, retries with exponential backoff + jitter up to MaxRetries.
//   - Respects context cancellation at every step.
func Do[T any](ctx context.Context, p *Policy, fn func() (T, error)) (T, error) {
	if p == nil {
		p = DefaultPolicy()
	}
	return doWithSleeper(ctx, p, fn, defaultSleeper)
}

func doWithSleeper[T any](ctx context.Context, p *Policy, fn func() (T, error), s sleeper) (T, error) {
	var zero T
	maxAttempts := p.MaxRetries + 1 // first call + retries

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check context before each attempt.
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		// Permanent errors are never retried.
		if permErr, ok := err.(*PermanentError); ok {
			return zero, permErr
		}

		// Flood wait errors: wait if within budget, otherwise fail.
		if fwErr, ok := err.(*FloodWaitError); ok {
			waitDuration := time.Duration(fwErr.Seconds) * time.Second
			if waitDuration > p.MaxFloodWait {
				return zero, fwErr
			}
			fmt.Fprintf(os.Stderr, "[tgs] FLOOD_WAIT: waiting %ds before retry (attempt %d/%d)\n",
				fwErr.Seconds, attempt, maxAttempts)

			if err := s.Sleep(ctx, waitDuration); err != nil {
				return zero, err
			}
			// Flood wait does NOT consume a retry attempt — restart loop
			// at the same attempt number.
			attempt--
			maxAttempts = p.MaxRetries + 1 // keep ceiling stable
			continue
		}

		// Transient error: if we haven't exhausted retries, backoff and retry.
		if attempt >= maxAttempts {
			return zero, err
		}

		delay := backoff(attempt, p.BaseDelay, p.MaxDelay)
		fmt.Fprintf(os.Stderr, "[tgs] retry: %v, backing off %v (attempt %d/%d)\n",
			err, delay, attempt, maxAttempts)

		if err := s.Sleep(ctx, delay); err != nil {
			return zero, err
		}
	}

	// Should not be reached, but satisfy the compiler.
	return zero, fmt.Errorf("retry: exhausted")
}

// backoff computes exponential backoff with full jitter.
// attempt is 1-based (first retry = attempt 1).
func backoff(attempt int, base, max time.Duration) time.Duration {
	exp := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(base) * exp)
	if delay > max {
		delay = max
	}
	// Full jitter: uniform random in [0, delay].
	if delay > 0 {
		delay = time.Duration(rand.Int64N(int64(delay)) + 1)
	}
	return delay
}
