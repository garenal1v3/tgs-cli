package resolve

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)


// mockAPI implements the API interface for testing.
type mockAPI struct {
	resolveUsername func(ctx context.Context, req *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error)
	resolvePhone    func(ctx context.Context, phone string) (*tg.ContactsResolvedPeer, error)
}

func (m *mockAPI) ContactsResolveUsername(ctx context.Context, req *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
	return m.resolveUsername(ctx, req)
}

func (m *mockAPI) ContactsResolvePhone(ctx context.Context, phone string) (*tg.ContactsResolvedPeer, error) {
	return m.resolvePhone(ctx, phone)
}

// makeUserResolved builds a ContactsResolvedPeer for a user with the given ID and access hash.
func makeUserResolved(userID, accessHash int64) *tg.ContactsResolvedPeer {
	u := &tg.User{ID: userID}
	u.SetAccessHash(accessHash)
	return &tg.ContactsResolvedPeer{
		Peer:  &tg.PeerUser{UserID: userID},
		Users: []tg.UserClass{u},
	}
}

func TestResolver_ResolveUsername(t *testing.T) {
	api := &mockAPI{
		resolveUsername: func(_ context.Context, req *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			if req.Username != "durov" {
				t.Errorf("unexpected username: %q", req.Username)
			}
			return makeUserResolved(123, 456), nil
		},
	}

	r := NewResolver(api, nil)
	peer, err := r.Resolve(context.Background(), "@durov")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	u, ok := peer.(*tg.InputPeerUser)
	if !ok {
		t.Fatalf("expected *tg.InputPeerUser, got %T", peer)
	}
	if u.UserID != 123 {
		t.Errorf("UserID = %d, want 123", u.UserID)
	}
	if u.AccessHash != 456 {
		t.Errorf("AccessHash = %d, want 456", u.AccessHash)
	}
}

func TestResolver_ResolvePhone(t *testing.T) {
	api := &mockAPI{
		resolvePhone: func(_ context.Context, phone string) (*tg.ContactsResolvedPeer, error) {
			if phone != "79001234567" {
				t.Errorf("unexpected phone: %q", phone)
			}
			return makeUserResolved(789, 101), nil
		},
	}

	r := NewResolver(api, nil)
	peer, err := r.Resolve(context.Background(), "+79001234567")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	u, ok := peer.(*tg.InputPeerUser)
	if !ok {
		t.Fatalf("expected *tg.InputPeerUser, got %T", peer)
	}
	if u.UserID != 789 {
		t.Errorf("UserID = %d, want 789", u.UserID)
	}
	if u.AccessHash != 101 {
		t.Errorf("AccessHash = %d, want 101", u.AccessHash)
	}
}

func TestResolver_ResolveMulti(t *testing.T) {
	var callCount atomic.Int32

	api := &mockAPI{
		resolveUsername: func(_ context.Context, req *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			callCount.Add(1)
			switch req.Username {
			case "alice":
				return makeUserResolved(1, 10), nil
			case "bob":
				return makeUserResolved(2, 20), nil
			case "charlie":
				return makeUserResolved(3, 30), nil
			default:
				t.Errorf("unexpected username: %q", req.Username)
				return makeUserResolved(0, 0), nil
			}
		},
	}

	r := NewResolver(api, nil)
	peers, err := r.ResolveMulti(context.Background(), []string{"@alice", "@bob", "@charlie"})
	if err != nil {
		t.Fatalf("ResolveMulti: %v", err)
	}

	if len(peers) != 3 {
		t.Fatalf("got %d peers, want 3", len(peers))
	}

	if got := callCount.Load(); got != 3 {
		t.Errorf("API call count = %d, want 3", got)
	}

	wantIDs := []int64{1, 2, 3}
	for i, peer := range peers {
		u, ok := peer.(*tg.InputPeerUser)
		if !ok {
			t.Errorf("peers[%d]: expected *tg.InputPeerUser, got %T", i, peer)
			continue
		}
		if u.UserID != wantIDs[i] {
			t.Errorf("peers[%d].UserID = %d, want %d", i, u.UserID, wantIDs[i])
		}
	}
}

func TestResolver_CacheHit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cache, err := NewPeerCache(dbPath)
	if err != nil {
		t.Fatalf("NewPeerCache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	// Pre-populate cache with entry for "durov".
	entry := CacheEntry{
		PeerType:        "user",
		ID:              777,
		AccessHash:      888,
		ResolvedAt:      time.Now().Unix(),
		SnapshotVersion: 1,
	}
	if err := cache.Store("durov", entry); err != nil {
		t.Fatalf("cache.Store: %v", err)
	}

	var callCount atomic.Int32

	api := &mockAPI{
		resolveUsername: func(_ context.Context, _ *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			callCount.Add(1)
			return makeUserResolved(777, 888), nil
		},
	}

	r := NewResolver(api, cache)
	peer, err := r.Resolve(context.Background(), "@durov")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if got := callCount.Load(); got != 0 {
		t.Errorf("API call count = %d, want 0 (should have used cache)", got)
	}

	u, ok := peer.(*tg.InputPeerUser)
	if !ok {
		t.Fatalf("expected *tg.InputPeerUser, got %T", peer)
	}
	if u.UserID != 777 {
		t.Errorf("UserID = %d, want 777", u.UserID)
	}
	if u.AccessHash != 888 {
		t.Errorf("AccessHash = %d, want 888", u.AccessHash)
	}
}

// TestResolver_ResolveID_LegacyCacheKeepsAccessHash verifies that an
// ID-keyed cache entry written before SnapshotVersion was introduced still
// yields an InputPeer carrying its cached AccessHash. Regression: without
// this, search commands break for any peer cached under an older binary
// (they'd resolve to a peer with AccessHash=0 → PEER_ID_INVALID at call
// time).
func TestResolver_ResolveID_LegacyCacheKeepsAccessHash(t *testing.T) {
	cache, err := NewPeerCache(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	// Legacy entry — no SnapshotVersion, no snapshot fields, but has access_hash.
	_ = cache.Store("id:1234567890", CacheEntry{
		PeerType: "channel", ID: 1234567890, AccessHash: 999, ResolvedAt: time.Now().Unix(),
	})

	r := NewResolver(nil, cache)
	peer, err := r.Resolve(context.Background(), "-1001234567890")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	ch, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		t.Fatalf("expected *tg.InputPeerChannel, got %T", peer)
	}
	if ch.AccessHash != 999 {
		t.Errorf("AccessHash = %d, want 999 (from legacy cache)", ch.AccessHash)
	}
	if ch.ChannelID != 1234567890 {
		t.Errorf("ChannelID = %d, want 1234567890", ch.ChannelID)
	}
}

func TestResolver_ResolveID_Positive(t *testing.T) {
	r := NewResolver(nil, nil)
	peer, err := r.Resolve(context.Background(), "42")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	u, ok := peer.(*tg.InputPeerUser)
	if !ok {
		t.Fatalf("expected *tg.InputPeerUser, got %T", peer)
	}
	if u.UserID != 42 {
		t.Errorf("UserID = %d, want 42", u.UserID)
	}
}

func TestResolver_ResolveID_Channel(t *testing.T) {
	r := NewResolver(nil, nil)
	// -1001234567890 should map to channel ID 1234567890
	peer, err := r.Resolve(context.Background(), "-1001234567890")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	ch, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		t.Fatalf("expected *tg.InputPeerChannel, got %T", peer)
	}
	if ch.ChannelID != 1234567890 {
		t.Errorf("ChannelID = %d, want 1234567890", ch.ChannelID)
	}
}

func TestResolver_ResolveID_Chat(t *testing.T) {
	r := NewResolver(nil, nil)
	peer, err := r.Resolve(context.Background(), "-123456")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	ch, ok := peer.(*tg.InputPeerChat)
	if !ok {
		t.Fatalf("expected *tg.InputPeerChat, got %T", peer)
	}
	if ch.ChatID != 123456 {
		t.Errorf("ChatID = %d, want 123456", ch.ChatID)
	}
}

func TestResolver_ResolveInvite(t *testing.T) {
	r := NewResolver(nil, nil)
	_, err := r.Resolve(context.Background(), "t.me/+abc123invite")
	if err == nil {
		t.Fatal("expected error for invite link, got nil")
	}
}

func TestResolver_ResolveWithMeta_ChannelSubscribed(t *testing.T) {
	ch := &tg.Channel{ID: 1234, Title: "Pub", Broadcast: true, Left: false}
	ch.SetAccessHash(99)
	ch.SetUsername("pubch")
	api := &mockAPI{
		resolveUsername: func(_ context.Context, _ *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			return &tg.ContactsResolvedPeer{
				Peer:  &tg.PeerChannel{ChannelID: 1234},
				Chats: []tg.ChatClass{ch},
			}, nil
		},
	}
	r := NewResolver(api, nil)
	_, meta, err := r.ResolveWithMeta(context.Background(), "@pubch")
	if err != nil {
		t.Fatalf("ResolveWithMeta: %v", err)
	}
	if meta.PeerType != "channel" {
		t.Errorf("PeerType = %q, want channel", meta.PeerType)
	}
	if meta.Subscribed == nil || *meta.Subscribed != true {
		t.Errorf("Subscribed = %v, want *true", meta.Subscribed)
	}
}

func TestResolver_ResolveWithMeta_ChannelLeft(t *testing.T) {
	ch := &tg.Channel{ID: 1234, Title: "Pub", Broadcast: true, Left: true}
	ch.SetAccessHash(99)
	ch.SetUsername("pubch")
	api := &mockAPI{
		resolveUsername: func(_ context.Context, _ *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			return &tg.ContactsResolvedPeer{
				Peer:  &tg.PeerChannel{ChannelID: 1234},
				Chats: []tg.ChatClass{ch},
			}, nil
		},
	}
	r := NewResolver(api, nil)
	_, meta, err := r.ResolveWithMeta(context.Background(), "@pubch")
	if err != nil {
		t.Fatalf("ResolveWithMeta: %v", err)
	}
	if meta.Subscribed == nil || *meta.Subscribed != false {
		t.Errorf("Subscribed = %v, want *false", meta.Subscribed)
	}
}

func TestResolver_ResolveWithMeta_CacheHitNoSubscribedInfo(t *testing.T) {
	cache, err := NewPeerCache(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	_ = cache.Store("cached", CacheEntry{
		PeerType: "channel", ID: 7, AccessHash: 11, ResolvedAt: time.Now().Unix(),
		SnapshotVersion: 1,
		// Snapshot fields intentionally empty: simulate a cached entry that
		// recorded routing info but had no extra channel metadata (e.g. partial
		// resolve). Subscribed should remain nil — not populated from absent
		// snapshot data.
	})

	r := NewResolver(nil, cache)
	_, meta, err := r.ResolveWithMeta(context.Background(), "@cached")
	if err != nil {
		t.Fatalf("ResolveWithMeta: %v", err)
	}
	if meta.PeerType != "channel" {
		t.Errorf("PeerType = %q", meta.PeerType)
	}
	if meta.Subscribed != nil {
		t.Errorf("Subscribed should be nil for empty-snapshot cache hits, got %v", *meta.Subscribed)
	}
}

func TestResolver_ResolveWithMeta_ChannelSnapshot(t *testing.T) {
	ch := &tg.Channel{
		ID: 1234, Title: "Pub", Broadcast: true, Verified: true, Left: false,
	}
	ch.SetAccessHash(99)
	ch.SetUsername("pubch")
	ch.SetParticipantsCount(5000)
	api := &mockAPI{
		resolveUsername: func(_ context.Context, _ *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			return &tg.ContactsResolvedPeer{
				Peer:  &tg.PeerChannel{ChannelID: 1234},
				Chats: []tg.ChatClass{ch},
			}, nil
		},
	}
	r := NewResolver(api, nil)
	_, meta, err := r.ResolveWithMeta(context.Background(), "@pubch")
	if err != nil {
		t.Fatalf("ResolveWithMeta: %v", err)
	}
	if meta.Snapshot == nil {
		t.Fatal("Snapshot is nil, want populated")
	}
	if meta.Snapshot.Title != "Pub" {
		t.Errorf("Snapshot.Title = %q, want Pub", meta.Snapshot.Title)
	}
	if meta.Snapshot.Username != "pubch" {
		t.Errorf("Snapshot.Username = %q, want pubch", meta.Snapshot.Username)
	}
	if meta.Snapshot.Access != "public" {
		t.Errorf("Snapshot.Access = %q, want public", meta.Snapshot.Access)
	}
	if meta.Snapshot.MembersCount != 5000 {
		t.Errorf("Snapshot.MembersCount = %d, want 5000", meta.Snapshot.MembersCount)
	}
	if !meta.Snapshot.Verified {
		t.Error("Snapshot.Verified = false, want true")
	}
	if !meta.Snapshot.Broadcast {
		t.Error("Snapshot.Broadcast = false, want true")
	}
}

func TestResolver_ResolveWithMeta_CacheHitReturnsSnapshot(t *testing.T) {
	cache, err := NewPeerCache(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	sub := true
	_ = cache.Store("cached", CacheEntry{
		PeerType: "channel", ID: 7, AccessHash: 11, ResolvedAt: time.Now().Unix(),
		SnapshotVersion: 1,
		Title:           "Cached Ch", Username: "cached", Access: "public",
		MembersCount: 42, Broadcast: true, Subscribed: &sub,
	})
	r := NewResolver(nil, cache)
	_, meta, err := r.ResolveWithMeta(context.Background(), "@cached")
	if err != nil {
		t.Fatalf("ResolveWithMeta: %v", err)
	}
	if meta.Snapshot == nil {
		t.Fatal("Snapshot is nil, want populated from cache")
	}
	if meta.Snapshot.Title != "Cached Ch" {
		t.Errorf("Snapshot.Title = %q, want Cached Ch", meta.Snapshot.Title)
	}
	if meta.Subscribed == nil || !*meta.Subscribed {
		t.Errorf("Subscribed = %v, want *true from cache", meta.Subscribed)
	}
}

func TestResolver_ResolveWithMeta_UserSnapshot(t *testing.T) {
	u := &tg.User{ID: 42, Bot: true, Verified: true}
	u.SetFirstName("MyBot")
	u.SetUsername("mybot")
	u.SetAccessHash(99)
	api := &mockAPI{
		resolveUsername: func(_ context.Context, _ *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			return &tg.ContactsResolvedPeer{Peer: &tg.PeerUser{UserID: 42}, Users: []tg.UserClass{u}}, nil
		},
	}
	r := NewResolver(api, nil)
	_, meta, err := r.ResolveWithMeta(context.Background(), "@mybot")
	if err != nil {
		t.Fatalf("ResolveWithMeta: %v", err)
	}
	if meta.Snapshot == nil {
		t.Fatal("Snapshot is nil, want populated")
	}
	if meta.Snapshot.FirstName != "MyBot" || !meta.Snapshot.IsBot || !meta.Snapshot.Verified {
		t.Errorf("snapshot user fields wrong: %+v", meta.Snapshot)
	}
}

func TestResolver_ResolveWithMeta_UserNoSubscribed(t *testing.T) {
	api := &mockAPI{
		resolveUsername: func(_ context.Context, _ *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
			return makeUserResolved(42, 99), nil
		},
	}
	r := NewResolver(api, nil)
	_, meta, err := r.ResolveWithMeta(context.Background(), "@alice")
	if err != nil {
		t.Fatalf("ResolveWithMeta: %v", err)
	}
	if meta.PeerType != "user" {
		t.Errorf("PeerType = %q, want user", meta.PeerType)
	}
	if meta.Subscribed != nil {
		t.Error("Subscribed should be nil for users")
	}
}
