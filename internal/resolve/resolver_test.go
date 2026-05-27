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
	resolvePhone   func(ctx context.Context, req *tg.ContactsResolvePhoneRequest) (*tg.ContactsResolvedPeer, error)
}

func (m *mockAPI) ContactsResolveUsername(ctx context.Context, req *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
	return m.resolveUsername(ctx, req)
}

func (m *mockAPI) ContactsResolvePhone(ctx context.Context, req *tg.ContactsResolvePhoneRequest) (*tg.ContactsResolvedPeer, error) {
	return m.resolvePhone(ctx, req)
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
		resolvePhone: func(_ context.Context, req *tg.ContactsResolvePhoneRequest) (*tg.ContactsResolvedPeer, error) {
			if req.Phone != "79001234567" {
				t.Errorf("unexpected phone: %q", req.Phone)
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
	defer cache.Close()

	// Pre-populate cache with entry for "durov".
	entry := CacheEntry{
		PeerType:   "user",
		ID:         777,
		AccessHash: 888,
		ResolvedAt: time.Now().Unix(),
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
