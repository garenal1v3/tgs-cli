package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gotd/td/tgerr"
)

// fakeSleeper records sleep calls and returns immediately (or simulates cancellation).
type fakeSleeper struct {
	sleeps []time.Duration
}

func (f *fakeSleeper) Sleep(_ context.Context, d time.Duration) error {
	f.sleeps = append(f.sleeps, d)
	return nil
}

func tinyPolicy() *Policy {
	return &Policy{
		MaxRetries:   3,
		MaxFloodWait: 60 * time.Second,
		BaseDelay:    time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
	}
}

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	p := tinyPolicy()
	s := &fakeSleeper{}

	result, err := doWithSleeper(context.Background(), p, func() (string, error) {
		return "ok", nil
	}, s)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("expected 'ok', got %q", result)
	}
	if len(s.sleeps) != 0 {
		t.Fatalf("expected no sleeps, got %d", len(s.sleeps))
	}
}

func TestDo_RetriesOnTransientError(t *testing.T) {
	p := tinyPolicy()
	s := &fakeSleeper{}
	calls := 0

	result, err := doWithSleeper(context.Background(), p, func() (int, error) {
		calls++
		if calls < 3 {
			return 0, errors.New("transient")
		}
		return 42, nil
	}, s)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	// Should have slept between retries (2 backoff sleeps).
	if len(s.sleeps) != 2 {
		t.Fatalf("expected 2 sleeps, got %d", len(s.sleeps))
	}
}

func TestDo_ExhaustsRetries(t *testing.T) {
	p := tinyPolicy()
	s := &fakeSleeper{}
	calls := 0

	_, err := doWithSleeper(context.Background(), p, func() (int, error) {
		calls++
		return 0, errors.New("always fails")
	}, s)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "always fails" {
		t.Fatalf("expected 'always fails', got %q", err.Error())
	}
	// 1 initial + 3 retries = 4 total calls
	if calls != 4 {
		t.Fatalf("expected 4 calls (1 + 3 retries), got %d", calls)
	}
	// 3 backoff sleeps (between attempts)
	if len(s.sleeps) != 3 {
		t.Fatalf("expected 3 sleeps, got %d", len(s.sleeps))
	}
}

func TestDo_RespectsContextCancellation(t *testing.T) {
	p := tinyPolicy()
	calls := 0

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := Do(ctx, p, func() (int, error) {
		calls++
		return 0, errors.New("should not reach")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected 0 calls, got %d", calls)
	}
}

func TestDo_WaitsOnFloodWait(t *testing.T) {
	p := tinyPolicy()
	s := &fakeSleeper{}
	calls := 0

	result, err := doWithSleeper(context.Background(), p, func() (string, error) {
		calls++
		if calls == 1 {
			return "", &FloodWaitError{Seconds: 1}
		}
		return "done", nil
	}, s)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "done" {
		t.Fatalf("expected 'done', got %q", result)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	// Should have slept for 1 second (flood wait).
	if len(s.sleeps) != 1 {
		t.Fatalf("expected 1 sleep, got %d", len(s.sleeps))
	}
	if s.sleeps[0] != time.Second {
		t.Fatalf("expected 1s sleep, got %v", s.sleeps[0])
	}
}

func TestDo_RejectsLongFloodWait(t *testing.T) {
	p := tinyPolicy()
	p.MaxFloodWait = 30 * time.Second
	s := &fakeSleeper{}

	_, err := doWithSleeper(context.Background(), p, func() (int, error) {
		return 0, &FloodWaitError{Seconds: 120}
	}, s)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var fwErr *FloodWaitError
	if !errors.As(err, &fwErr) {
		t.Fatalf("expected FloodWaitError, got %T: %v", err, err)
	}
	if fwErr.Seconds != 120 {
		t.Fatalf("expected 120 seconds, got %d", fwErr.Seconds)
	}
	// Should not have slept at all.
	if len(s.sleeps) != 0 {
		t.Fatalf("expected 0 sleeps, got %d", len(s.sleeps))
	}
}

func TestDo_NoPermanentRetry(t *testing.T) {
	p := tinyPolicy()
	s := &fakeSleeper{}
	calls := 0

	_, err := doWithSleeper(context.Background(), p, func() (int, error) {
		calls++
		return 0, &PermanentError{Err: errors.New("CHANNEL_PRIVATE")}
	}, s)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var permErr *PermanentError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected PermanentError, got %T: %v", err, err)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call (no retries), got %d", calls)
	}
	if len(s.sleeps) != 0 {
		t.Fatalf("expected 0 sleeps, got %d", len(s.sleeps))
	}
}

func TestClassifyError_FloodWait(t *testing.T) {
	err := tgerr.New(420, "FLOOD_WAIT_5")
	classified := ClassifyError(err)

	var fwErr *FloodWaitError
	if !errors.As(classified, &fwErr) {
		t.Fatalf("expected FloodWaitError, got %T: %v", classified, classified)
	}
	if fwErr.Seconds != 5 {
		t.Fatalf("expected 5 seconds, got %d", fwErr.Seconds)
	}
}

func TestClassifyError_Permanent(t *testing.T) {
	err := tgerr.New(400, "CHANNEL_PRIVATE")
	classified := ClassifyError(err)

	var permErr *PermanentError
	if !errors.As(classified, &permErr) {
		t.Fatalf("expected PermanentError, got %T: %v", classified, classified)
	}
}

// TestClassifyError_UserIDInvalidIsPermanent guards against regressing the
// retry list: USER_ID_INVALID is not a transient error and must NOT trigger
// the exponential-backoff loop.
func TestClassifyError_UserIDInvalidIsPermanent(t *testing.T) {
	for _, code := range []string{"USER_ID_INVALID", "INPUT_USER_DEACTIVATED"} {
		err := tgerr.New(400, code)
		classified := ClassifyError(err)
		var permErr *PermanentError
		if !errors.As(classified, &permErr) {
			t.Errorf("%s: expected PermanentError, got %T: %v", code, classified, classified)
		}
	}
}

func TestClassifyError_Transient(t *testing.T) {
	err := tgerr.New(500, "RPC_CALL_FAIL")
	classified := ClassifyError(err)

	// Should pass through unchanged — not a FloodWaitError or PermanentError.
	var fwErr *FloodWaitError
	if errors.As(classified, &fwErr) {
		t.Fatal("should not be FloodWaitError")
	}
	var permErr *PermanentError
	if errors.As(classified, &permErr) {
		t.Fatal("should not be PermanentError")
	}

	// Should still be a tgerr.Error.
	var rpcErr *tgerr.Error
	if !errors.As(classified, &rpcErr) {
		t.Fatalf("expected tgerr.Error, got %T: %v", classified, classified)
	}
	if rpcErr.Code != 500 {
		t.Fatalf("expected code 500, got %d", rpcErr.Code)
	}
}
