package sources

import (
	"path/filepath"
	"testing"
)

func TestSourceStatsCache_PutAndGet(t *testing.T) {
	c, err := NewSourceStatsCache(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatalf("NewSourceStatsCache: %v", err)
	}
	defer func() { _ = c.Close() }()

	payload := &CachedPayload{
		MembersCount: 100,
		Description:  "hi",
		Stats: &Stats{
			TotalMessages: 5000,
			Messages24h:   2,
			FirstMessage:  &FirstMessage{ID: 1, Date: "2020-01-01T00:00:00Z"},
		},
	}
	if err := c.Put(-1001, 4242, payload); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, found, err := c.Get(-1001, 4242)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if got.MembersCount != 100 || got.Description != "hi" {
		t.Errorf("payload mismatch: %+v", got)
	}
	if got.Stats == nil || got.Stats.TotalMessages != 5000 {
		t.Errorf("stats mismatch: %+v", got.Stats)
	}
}

func TestSourceStatsCache_StaleByLastMessageID(t *testing.T) {
	c, err := NewSourceStatsCache(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatalf("NewSourceStatsCache: %v", err)
	}
	defer func() { _ = c.Close() }()

	if err := c.Put(-1001, 100, &CachedPayload{MembersCount: 10}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Same last_message_id -> hit
	_, found, err := c.Get(-1001, 100)
	if err != nil || !found {
		t.Fatalf("expected hit on same id, found=%v err=%v", found, err)
	}

	// Different last_message_id -> miss (stale)
	_, found, err = c.Get(-1001, 101)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Error("expected miss when last_message_id differs")
	}
}

func TestSourceStatsCache_Missing(t *testing.T) {
	c, err := NewSourceStatsCache(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatalf("NewSourceStatsCache: %v", err)
	}
	defer func() { _ = c.Close() }()

	_, found, err := c.Get(-999, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Error("expected found=false for missing key")
	}
}
