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
	defer cache.Close()

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
	defer cache.Close()

	_, found, err := cache.Load("@nonexistent")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if found {
		t.Fatal("Load: expected found=false for missing key, got true")
	}
}

func TestPeerCache_TTLExpired(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	cache, err := NewPeerCache(dbPath)
	if err != nil {
		t.Fatalf("NewPeerCache: %v", err)
	}
	defer cache.Close()

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
