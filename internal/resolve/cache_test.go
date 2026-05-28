package resolve

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPeerCache_StoreAndLoad(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	cache, err := NewPeerCache(dbPath)
	if err != nil {
		t.Fatalf("NewPeerCache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	want := CacheEntry{
		PeerType:   "channel",
		ID:         1234567890,
		AccessHash: -9876543210,
		ResolvedAt: time.Now().Unix(),
	}

	if err := cache.Store("@testchannel", want); err != nil {
		t.Fatalf("Store: %v", err)
	}

	got, found, err := cache.Load("@testchannel")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !found {
		t.Fatal("Load: expected found=true, got false")
	}
	if got.PeerType != want.PeerType {
		t.Errorf("PeerType = %q, want %q", got.PeerType, want.PeerType)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
	if got.AccessHash != want.AccessHash {
		t.Errorf("AccessHash = %d, want %d", got.AccessHash, want.AccessHash)
	}
	if got.ResolvedAt != want.ResolvedAt {
		t.Errorf("ResolvedAt = %d, want %d", got.ResolvedAt, want.ResolvedAt)
	}
}

func TestPeerCache_LoadMissing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	cache, err := NewPeerCache(dbPath)
	if err != nil {
		t.Fatalf("NewPeerCache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	_, found, err := cache.Load("@nonexistent")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if found {
		t.Fatal("Load: expected found=false for missing key, got true")
	}
}

func TestPeerCache_StoreAndLoadSnapshot(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	cache, err := NewPeerCache(dbPath)
	if err != nil {
		t.Fatalf("NewPeerCache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	want := CacheEntry{
		PeerType:        "channel",
		ID:              42,
		AccessHash:      1,
		ResolvedAt:      time.Now().Unix(),
		SnapshotVersion: 1,
		Title:           "Durov's Channel",
		Username:     "durov",
		Access:       "public",
		MembersCount: 1234567,
		Verified:     true,
		HasTopics:    false,
		Gigagroup:    false,
		Broadcast:    true,
		Subscribed:   boolPtr(true),
	}
	if err := cache.Store("@durov", want); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, found, err := cache.Load("@durov")
	if err != nil || !found {
		t.Fatalf("Load: err=%v found=%v", err, found)
	}
	if got.Title != "Durov's Channel" || got.Username != "durov" || got.Access != "public" {
		t.Errorf("snapshot strings lost: %+v", got)
	}
	if got.MembersCount != 1234567 {
		t.Errorf("MembersCount = %d, want 1234567", got.MembersCount)
	}
	if !got.Verified || !got.Broadcast {
		t.Errorf("bool fields lost: verified=%v broadcast=%v", got.Verified, got.Broadcast)
	}
	if got.Subscribed == nil || !*got.Subscribed {
		t.Errorf("Subscribed = %v, want *true", got.Subscribed)
	}
}

func boolPtr(v bool) *bool { return &v }

func TestPeerCache_TTLExpired(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	cache, err := NewPeerCache(dbPath)
	if err != nil {
		t.Fatalf("NewPeerCache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	// Set TTL to 1 second.
	cache.TTL = 1 * time.Second

	entry := CacheEntry{
		PeerType:   "user",
		ID:         42,
		AccessHash: 100,
		ResolvedAt: time.Now().Add(-2 * time.Second).Unix(), // 2 seconds ago
	}

	if err := cache.Store("@expired", entry); err != nil {
		t.Fatalf("Store: %v", err)
	}

	_, found, err := cache.Load("@expired")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if found {
		t.Fatal("Load: expected found=false for expired entry, got true")
	}
}
