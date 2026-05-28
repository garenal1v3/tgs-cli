package sources

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gotd/td/tg"

	"github.com/searchtgcli/tgs/internal/resolve"
)

func TestService_List_BasicFiltering(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			channel := &tg.Channel{ID: 1, Title: "Ch", Broadcast: true}
			channel.SetAccessHash(1)
			group := &tg.Chat{ID: 2, Title: "Grp"}
			user := &tg.User{ID: 3}
			user.SetFirstName("Alice")
			user.SetAccessHash(3)
			return &tg.MessagesDialogs{
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10},
					&tg.Dialog{Peer: &tg.PeerChat{ChatID: 2}, TopMessage: 20},
					&tg.Dialog{Peer: &tg.PeerUser{UserID: 3}, TopMessage: 30},
				},
				Messages: []tg.MessageClass{
					&tg.Message{ID: 10, Date: 1700000000},
					&tg.Message{ID: 20, Date: 1700000000},
					&tg.Message{ID: 30, Date: 1700000000},
				},
				Chats: []tg.ChatClass{channel, group},
				Users: []tg.UserClass{user},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)

	got, err := s.List(context.Background(), ListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Sources) != 3 {
		t.Errorf("len = %d, want 3", len(got.Sources))
	}

	// Filter
	got, err = s.List(context.Background(), ListRequest{Types: []string{"channel"}})
	if err != nil {
		t.Fatalf("List filter: %v", err)
	}
	if len(got.Sources) != 1 || got.Sources[0].Type != "channel" {
		t.Errorf("filtered = %+v", got.Sources)
	}
}

func TestService_List_WithStats_CacheHitAvoidsAPI(t *testing.T) {
	weekOld := int(time.Now().Add(-8 * 24 * time.Hour).Unix())

	apiCalls := 0
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "Old", Broadcast: true}
			ch.SetAccessHash(1)
			return &tg.MessagesDialogs{
				Dialogs:  []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 500}},
				Messages: []tg.MessageClass{&tg.Message{ID: 500, Date: weekOld}},
				Chats:    []tg.ChatClass{ch},
			}, nil
		},
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			apiCalls++
			return &tg.MessagesChatFull{FullChat: &tg.ChannelFull{ID: 1, ParticipantsCount: 5}}, nil
		},
		search: func(_ context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			apiCalls++
			return &tg.MessagesMessagesSlice{Count: 1}, nil
		},
		getHistory: func(_ context.Context, _ *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			apiCalls++
			return &tg.MessagesMessagesSlice{Messages: nil}, nil
		},
	}

	cache, err := NewSourceStatsCache(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cache.Close() }()
	s := New(api, nil, nil, cache, 0)

	// First run: live (cold cache).
	if _, err := s.List(context.Background(), ListRequest{WithStats: true}); err != nil {
		t.Fatalf("first List: %v", err)
	}
	firstRunCalls := apiCalls
	if firstRunCalls == 0 {
		t.Fatal("expected API calls on cold run")
	}

	// Second run: warm cache, same last_message_id. No new expensive calls.
	apiCalls = 0
	if _, err := s.List(context.Background(), ListRequest{WithStats: true}); err != nil {
		t.Fatalf("second List: %v", err)
	}
	if apiCalls != 0 {
		t.Errorf("warm cache should make 0 expensive calls, got %d", apiCalls)
	}
}

func TestService_List_WithStats_ActiveSourceSkipsCache(t *testing.T) {
	freshDate := int(time.Now().Unix())

	apiCalls := 0
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "Fresh", Broadcast: true}
			ch.SetAccessHash(1)
			return &tg.MessagesDialogs{
				Dialogs:  []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 500}},
				Messages: []tg.MessageClass{&tg.Message{ID: 500, Date: freshDate}},
				Chats:    []tg.ChatClass{ch},
			}, nil
		},
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			apiCalls++
			return &tg.MessagesChatFull{FullChat: &tg.ChannelFull{ID: 1}}, nil
		},
		search: func(_ context.Context, _ *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			apiCalls++
			return &tg.MessagesMessagesSlice{Count: 1}, nil
		},
		getHistory: func(_ context.Context, _ *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			apiCalls++
			return &tg.MessagesMessagesSlice{}, nil
		},
	}
	cache, err := NewSourceStatsCache(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cache.Close() }()
	s := New(api, nil, nil, cache, 0)

	// Cold run.
	_, _ = s.List(context.Background(), ListRequest{WithStats: true})
	apiCalls = 0
	// Warm run: source is "fresh" -> always live -> still hits API.
	_, _ = s.List(context.Background(), ListRequest{WithStats: true})
	if apiCalls == 0 {
		t.Error("active source should bypass cache and hit API again")
	}
}

func TestService_Inspect_Subscribed(t *testing.T) {
	api := &mockAPI{
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			return &tg.MessagesChatFull{FullChat: &tg.ChannelFull{ID: 1, About: "hi"}}, nil
		},
		search: func(_ context.Context, _ *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			return &tg.MessagesMessagesSlice{Count: 7}, nil
		},
		getHistory: func(_ context.Context, _ *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			return &tg.MessagesMessagesSlice{Messages: nil}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	src, err := s.inspectKnown(context.Background(), Source{
		ID: -1000000000001, Type: "channel",
	}, &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1}, false)
	if err != nil {
		t.Fatalf("inspectKnown: %v", err)
	}
	if src.Stats == nil || src.Stats.TotalMessages != 7 {
		t.Errorf("Stats = %+v", src.Stats)
	}
	if src.Subscribed == nil || *src.Subscribed != true {
		t.Errorf("Subscribed = %+v, want true", src.Subscribed)
	}
}

// TestService_Inspect_NotSubscribed_FetchesFullSkipsStats verifies that
// inspect of an unsubscribed public channel still pulls description / members
// via channels.getFullChannel (which works without membership for public
// channels), but skips messages.search / messages.getHistory (which require
// membership and would just produce empty stats).
func TestService_Inspect_NotSubscribed_FetchesFullSkipsStats(t *testing.T) {
	statsCalled := false
	api := &mockAPI{
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			full := &tg.ChannelFull{ID: 1, About: "official", ParticipantsCount: 9999}
			return &tg.MessagesChatFull{FullChat: full}, nil
		},
		search: func(_ context.Context, _ *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			statsCalled = true
			return &tg.MessagesMessagesSlice{}, nil
		},
		getHistory: func(_ context.Context, _ *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			statsCalled = true
			return &tg.MessagesMessagesSlice{}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	notSubscribed := false
	src, err := s.inspectKnownWithSubscription(context.Background(), Source{
		ID: -1000000000001, Type: "channel", Title: "Pre-set",
	}, &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1}, false, &notSubscribed)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if statsCalled {
		t.Error("messages.search / getHistory should be skipped when subscribed=*false")
	}
	if src.Subscribed == nil || *src.Subscribed != false {
		t.Errorf("Subscribed = %+v, want *false", src.Subscribed)
	}
	if src.Stats != nil {
		t.Errorf("Stats should be nil, got %+v", src.Stats)
	}
	if src.Description != "official" || src.MembersCount != 9999 {
		t.Errorf("expected fetchFull-supplied description/members, got %+v", src)
	}
}

// TestService_Inspect_NotSubscribed_PrivateGracefulOnFetchFullFail verifies
// that when fetchFull fails for an unsubscribed channel (e.g. CHANNEL_PRIVATE
// for a private channel), the seed Source is returned intact, Subscribed
// stays *false, and a StatsError code is recorded — but Stats remain nil.
func TestService_Inspect_NotSubscribed_PrivateGracefulOnFetchFullFail(t *testing.T) {
	api := &mockAPI{
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			return nil, errors.New("rpc error code 400: CHANNEL_PRIVATE (caused by ...)")
		},
	}
	s := New(api, nil, nil, nil, 0)
	no := false
	src, err := s.inspectKnownWithSubscription(context.Background(), Source{
		ID: -1000000000001, Type: "channel", Title: "Seed Title",
	}, &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1}, false, &no)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if src.Title != "Seed Title" {
		t.Errorf("Seed title lost: %q", src.Title)
	}
	if src.Subscribed == nil || *src.Subscribed != false {
		t.Errorf("Subscribed = %+v, want *false", src.Subscribed)
	}
	if src.StatsError != "CHANNEL_PRIVATE" {
		t.Errorf("StatsError = %q, want CHANNEL_PRIVATE", src.StatsError)
	}
	if src.Stats != nil {
		t.Error("Stats should be nil for unsubscribed channel")
	}
}

func TestService_Inspect_UnknownSubscribed_FallsBackToErrorCode(t *testing.T) {
	api := &mockAPI{
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			return nil, errors.New("rpc error code 400: CHANNEL_PRIVATE (caused by ...)")
		},
	}
	s := New(api, nil, nil, nil, 0)
	src, err := s.inspectKnownWithSubscription(context.Background(), Source{
		ID: -1000000000001, Type: "channel",
	}, &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1}, false, nil)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if src.Subscribed == nil || *src.Subscribed != false {
		t.Errorf("Subscribed = %+v, want *false (inferred from CHANNEL_PRIVATE)", src.Subscribed)
	}
	if src.StatsError != "CHANNEL_PRIVATE" {
		t.Errorf("StatsError = %q", src.StatsError)
	}
}

// TestService_List_FilterBeforeLimit verifies that --type filtering is applied
// per page BEFORE the limit check, so limit=2 with type=bot correctly returns
// bots even when the first page dialogs are non-bot.
//
// The mock returns pageSize (2) items on page 1 (both channels, no bots) and
// a bot on page 2 — simulating a Telegram account where the first N dialogs
// are all channels and bots appear later.
//
// Without the fix, after page 1 len(all)=2 >= limit=2 → loop breaks → 0 bots.
// With the fix, page 1 is filtered first → len(all)=0 < 2 → loop continues
// → page 2 finds the bot → 1 bot returned.
func TestService_List_FilterBeforeLimit(t *testing.T) {
	callCount := 0
	api := &mockAPI{
		getDialogs: func(_ context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			callCount++
			switch callCount {
			case 1:
				// Return 2 channels (filling the pageSize of 2). Count=10 so
				// len(dialogs)=2 >= limit=2 → cursor is generated → next page.
				ch1 := &tg.Channel{ID: 1, Title: "Ch1", Broadcast: true}
				ch1.SetAccessHash(1)
				ch2 := &tg.Channel{ID: 2, Title: "Ch2", Broadcast: true}
				ch2.SetAccessHash(2)
				return &tg.MessagesDialogsSlice{
					Count: 10,
					Dialogs: []tg.DialogClass{
						&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10},
						&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}, TopMessage: 9},
					},
					Messages: []tg.MessageClass{
						&tg.Message{ID: 10, Date: 1700000010},
						&tg.Message{ID: 9, Date: 1700000009},
					},
					Chats: []tg.ChatClass{ch1, ch2},
					Users: []tg.UserClass{},
				}, nil
			default:
				// Page 2: return the bot. Use MessagesDialogs (no cursor) to
				// signal end of list.
				bot := &tg.User{ID: 10, Bot: true}
				bot.SetFirstName("MyBot")
				bot.SetAccessHash(10)
				return &tg.MessagesDialogs{
					Dialogs:  []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerUser{UserID: 10}, TopMessage: 3}},
					Messages: []tg.MessageClass{&tg.Message{ID: 3, Date: 1700000000}},
					Chats:    []tg.ChatClass{},
					Users:    []tg.UserClass{bot},
				}, nil
			}
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, err := s.List(context.Background(), ListRequest{Types: []string{"bot"}, Limit: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Sources) != 1 {
		t.Fatalf("expected 1 bot, got %d: %+v", len(got.Sources), got.Sources)
	}
	if got.Sources[0].Type != "bot" {
		t.Errorf("Source.Type = %q, want bot", got.Sources[0].Type)
	}
	// Total must still reflect server-reported count (pre-filter).
	if got.Total != 10 {
		t.Errorf("Total = %d, want 10 (server-reported, not filtered)", got.Total)
	}
}

// TestService_Inspect_SelfPeerNoStatsError verifies that inspecting Saved
// Messages ("-") returns valid stats and no stats_error, even though the
// peer is *tg.InputPeerSelf (not *tg.InputPeerUser).
func TestService_Inspect_SelfPeerNoStatsError(t *testing.T) {
	const selfID int64 = 12345

	api := &mockAPI{
		getFullUser: func(_ context.Context, u tg.InputUserClass) (*tg.UsersUserFull, error) {
			// Accept both InputUserSelf (Saved Messages path) and InputUser.
			switch u.(type) {
			case *tg.InputUserSelf, *tg.InputUser:
			default:
				t.Errorf("unexpected InputUserClass type: %T", u)
			}
			full := &tg.UserFull{ID: selfID}
			full.SetAbout("myself")
			return &tg.UsersUserFull{FullUser: *full}, nil
		},
		search: func(_ context.Context, _ *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			return &tg.MessagesMessagesSlice{Count: 42}, nil
		},
		getHistory: func(_ context.Context, _ *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			return &tg.MessagesMessagesSlice{Messages: nil}, nil
		},
	}
	s := New(api, nil, nil, nil, selfID)
	src, err := s.inspectKnownWithSubscription(
		context.Background(),
		Source{ID: selfID, Type: "user", Saved: true, Title: "Saved Messages"},
		&tg.InputPeerSelf{},
		false,
		func() *bool { v := true; return &v }(),
	)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if src.StatsError != "" {
		t.Errorf("StatsError = %q, want empty", src.StatsError)
	}
	if src.Stats == nil {
		t.Fatal("Stats is nil, want non-nil")
	}
	if src.Stats.TotalMessages != 42 {
		t.Errorf("Stats.TotalMessages = %d, want 42", src.Stats.TotalMessages)
	}
	if !src.Saved {
		t.Error("Saved should be true")
	}
}

// TestService_List_TotalAtLeastReturnedCount verifies that when Telegram's
// reported MessagesDialogsSlice.Count is smaller than the number of dialogs
// we actually accumulated, the response uses the larger value — so
// `total >= len(sources)` always holds before filtering. Fixes P0-8.
//
// Telegram occasionally returns a Count smaller than the dialogs it actually
// shipped in the same page (race between the count cache and the slice
// pagination). We model that exact inconsistency here.
func TestService_List_TotalAtLeastReturnedCount(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch1 := &tg.Channel{ID: 1, Title: "C1", Broadcast: true}
			ch1.SetAccessHash(1)
			ch2 := &tg.Channel{ID: 2, Title: "C2", Broadcast: true}
			ch2.SetAccessHash(2)
			ch3 := &tg.Channel{ID: 3, Title: "C3", Broadcast: true}
			ch3.SetAccessHash(3)
			ch4 := &tg.Channel{ID: 4, Title: "C4", Broadcast: true}
			ch4.SetAccessHash(4)
			// Stale server count: claims 2 but ships 4. hasMore must stay
			// false (len(dialogs)=4 < pageSize=100) so we don't request a
			// next page.
			return &tg.MessagesDialogsSlice{
				Count: 2,
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10},
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}, TopMessage: 9},
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 3}, TopMessage: 8},
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 4}, TopMessage: 7},
				},
				Messages: []tg.MessageClass{
					&tg.Message{ID: 10, Date: 1700000010},
					&tg.Message{ID: 9, Date: 1700000009},
					&tg.Message{ID: 8, Date: 1700000008},
					&tg.Message{ID: 7, Date: 1700000007},
				},
				Chats: []tg.ChatClass{ch1, ch2, ch3, ch4},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, err := s.List(context.Background(), ListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Sources) != 4 {
		t.Fatalf("sources = %d, want 4", len(got.Sources))
	}
	if got.Total < len(got.Sources) {
		t.Errorf("Total = %d, want >= %d (len sources)", got.Total, len(got.Sources))
	}
	if got.Total != 4 {
		t.Errorf("Total = %d, want 4 (max of server count and totalSeen)", got.Total)
	}
}

func TestBuildInspectSeed_ChannelSnapshot(t *testing.T) {
	snap := &resolve.PeerSnapshot{
		Title: "Pub", Username: "pubch", Access: "public",
		MembersCount: 5000, Verified: true, Broadcast: true,
	}
	got := buildInspectSeed(&tg.InputPeerChannel{ChannelID: 1234, AccessHash: 99}, snap)
	if got.Type != "channel" {
		t.Errorf("Type = %q, want channel", got.Type)
	}
	if got.Title != "Pub" {
		t.Errorf("Title = %q, want Pub", got.Title)
	}
	if got.Username != "pubch" {
		t.Errorf("Username = %q, want pubch", got.Username)
	}
	if got.Access != "public" {
		t.Errorf("Access = %q, want public", got.Access)
	}
	if got.MembersCount != 5000 {
		t.Errorf("MembersCount = %d, want 5000", got.MembersCount)
	}
	if !got.Verified {
		t.Error("Verified should be true")
	}
}

func TestBuildInspectSeed_SupergroupSnapshot(t *testing.T) {
	snap := &resolve.PeerSnapshot{
		Title: "Group", Username: "grp", Access: "public",
		Broadcast: false, HasTopics: true,
	}
	got := buildInspectSeed(&tg.InputPeerChannel{ChannelID: 200}, snap)
	if got.Type != "supergroup" {
		t.Errorf("Type = %q, want supergroup", got.Type)
	}
	if !got.HasTopics {
		t.Error("HasTopics should be true")
	}
}

func TestBuildInspectSeed_UserSnapshot(t *testing.T) {
	snap := &resolve.PeerSnapshot{FirstName: "Alice", Username: "alice", IsBot: false, Verified: true}
	got := buildInspectSeed(&tg.InputPeerUser{UserID: 42, AccessHash: 1}, snap)
	if got.Type != "user" {
		t.Errorf("Type = %q, want user", got.Type)
	}
	if got.FirstName != "Alice" {
		t.Errorf("FirstName = %q", got.FirstName)
	}
	if got.Username != "alice" {
		t.Errorf("Username = %q", got.Username)
	}
}

func TestBuildInspectSeed_BotSnapshot(t *testing.T) {
	snap := &resolve.PeerSnapshot{FirstName: "MyBot", Username: "mybot", IsBot: true}
	got := buildInspectSeed(&tg.InputPeerUser{UserID: 10}, snap)
	if got.Type != "bot" {
		t.Errorf("Type = %q, want bot (IsBot=true)", got.Type)
	}
}

// TestBuildInspectSeed_SelfByPhone marks the seed as Saved Messages when
// the resolved peer is the authenticated user themselves (covers P2-17:
// inspect +<own-phone> should behave like inspect @me).
func TestBuildInspectSeed_SelfByPhone(t *testing.T) {
	const selfID int64 = 42
	snap := &resolve.PeerSnapshot{FirstName: "Me", Phone: "79001234567"}
	got := buildInspectSeed(&tg.InputPeerUser{UserID: selfID, AccessHash: 1}, snap)
	got = markSelf(got, selfID)
	if !got.Saved {
		t.Error("expected Saved=true when seed ID matches selfID")
	}
	if got.Title != "Saved Messages" {
		t.Errorf("Title = %q, want Saved Messages", got.Title)
	}
}

// TestBuildInspectSeed_OtherUserNotSaved verifies markSelf is a no-op for
// peers that aren't the authenticated user.
func TestBuildInspectSeed_OtherUserNotSaved(t *testing.T) {
	const selfID int64 = 42
	got := buildInspectSeed(&tg.InputPeerUser{UserID: 99}, &resolve.PeerSnapshot{FirstName: "Other"})
	got = markSelf(got, selfID)
	if got.Saved {
		t.Error("Saved should remain false for non-self user")
	}
}

func TestBuildInspectSeed_NilSnapshot_ChannelStillTyped(t *testing.T) {
	got := buildInspectSeed(&tg.InputPeerChannel{ChannelID: 1}, nil)
	if got.Type != "channel" {
		t.Errorf("Type = %q, want channel even without snapshot (default broadcast=false=supergroup unless snapshot says otherwise — but with nil we default to channel)", got.Type)
	}
	// We can't determine channel vs supergroup without a snapshot — default
	// to "channel" since it's the more common subscribe target.
	if got.Title != "" || got.Username != "" {
		t.Errorf("expected empty Title/Username with nil snapshot, got %+v", got)
	}
}

// TestService_List_ArchivedMergesFolders verifies that --archived returns
// dialogs from BOTH the main folder (0) and the archive folder (1), not just
// the archive. Telegram's messages.getDialogs reads one folder at a time, so
// List must walk folder 0 first, then folder 1.
func TestService_List_ArchivedMergesFolders(t *testing.T) {
	var foldersSeen []int
	api := &mockAPI{
		getDialogs: func(_ context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			folder, _ := req.GetFolderID()
			foldersSeen = append(foldersSeen, folder)
			switch folder {
			case 0:
				ch := &tg.Channel{ID: 1, Title: "MainCh", Broadcast: true}
				ch.SetAccessHash(1)
				return &tg.MessagesDialogs{
					Dialogs:  []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10}},
					Messages: []tg.MessageClass{&tg.Message{ID: 10, Date: 1700000010}},
					Chats:    []tg.ChatClass{ch},
				}, nil
			case 1:
				ch := &tg.Channel{ID: 2, Title: "ArchivedCh", Broadcast: true}
				ch.SetAccessHash(2)
				d := &tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}, TopMessage: 5}
				d.SetFolderID(1)
				return &tg.MessagesDialogs{
					Dialogs:  []tg.DialogClass{d},
					Messages: []tg.MessageClass{&tg.Message{ID: 5, Date: 1700000005}},
					Chats:    []tg.ChatClass{ch},
				}, nil
			}
			return &tg.MessagesDialogs{}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)

	got, err := s.List(context.Background(), ListRequest{Archived: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(foldersSeen) != 2 || foldersSeen[0] != 0 || foldersSeen[1] != 1 {
		t.Errorf("foldersSeen = %v, want [0 1]", foldersSeen)
	}
	if len(got.Sources) != 2 {
		t.Fatalf("sources = %d, want 2 (main + archived)", len(got.Sources))
	}
	titles := []string{got.Sources[0].Title, got.Sources[1].Title}
	if titles[0] != "MainCh" || titles[1] != "ArchivedCh" {
		t.Errorf("titles = %v, want [MainCh ArchivedCh]", titles)
	}
	// The archived source must carry Archived=true (derived from FolderID=1 on
	// the dialog).
	if !got.Sources[1].Archived {
		t.Error("archived source must have Archived=true")
	}
	// No more pages on either folder → no cursor.
	if got.Cursor != "" {
		t.Errorf("Cursor = %q, want empty (both folders exhausted)", got.Cursor)
	}
}

// TestService_List_NotArchived_DoesNotWalkArchive sanity-checks that the
// default (Archived=false) request still touches only folder 0 — the new
// folder-walking logic must not fire when --archived is absent.
func TestService_List_NotArchived_DoesNotWalkArchive(t *testing.T) {
	var foldersSeen []int
	api := &mockAPI{
		getDialogs: func(_ context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			folder, _ := req.GetFolderID()
			foldersSeen = append(foldersSeen, folder)
			ch := &tg.Channel{ID: 1, Title: "Ch", Broadcast: true}
			ch.SetAccessHash(1)
			return &tg.MessagesDialogs{
				Dialogs:  []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10}},
				Messages: []tg.MessageClass{&tg.Message{ID: 10, Date: 1700000010}},
				Chats:    []tg.ChatClass{ch},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	if _, err := s.List(context.Background(), ListRequest{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(foldersSeen) != 1 || foldersSeen[0] != 0 {
		t.Errorf("foldersSeen = %v, want [0]", foldersSeen)
	}
}

// TestService_List_ArchivedCursorResumesFolder1 verifies that when a previous
// page exhausted folder 0 and a cursor was emitted pointing at folder 1, a
// follow-up call with that cursor reads only folder 1 (and doesn't restart
// folder 0).
func TestService_List_ArchivedCursorResumesFolder1(t *testing.T) {
	var foldersSeen []int
	api := &mockAPI{
		getDialogs: func(_ context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			folder, _ := req.GetFolderID()
			foldersSeen = append(foldersSeen, folder)
			ch := &tg.Channel{ID: 2, Title: "Arch", Broadcast: true}
			ch.SetAccessHash(2)
			d := &tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}, TopMessage: 5}
			d.SetFolderID(1)
			return &tg.MessagesDialogs{
				Dialogs:  []tg.DialogClass{d},
				Messages: []tg.MessageClass{&tg.Message{ID: 5, Date: 1700000005}},
				Chats:    []tg.ChatClass{ch},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)

	resumeCursor := (&Cursor{Folder: 1}).Encode()
	got, err := s.List(context.Background(), ListRequest{Archived: true, Cursor: resumeCursor})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(foldersSeen) != 1 || foldersSeen[0] != 1 {
		t.Errorf("foldersSeen = %v, want [1] (cursor must resume on folder 1)", foldersSeen)
	}
	if len(got.Sources) != 1 || got.Sources[0].Title != "Arch" {
		t.Errorf("sources = %+v, want one archived source", got.Sources)
	}
}

// TestService_List_ArchivedLimitCrossFolderEmitsCursor checks that --archived
// with a --limit that lands mid-archive emits a cursor pointing at the last
// item we kept, so a follow-up call resumes in the archive folder instead of
// silently dropping the unread tail.
func TestService_List_ArchivedLimitCrossFolderEmitsCursor(t *testing.T) {
	calls := 0
	api := &mockAPI{
		getDialogs: func(_ context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			calls++
			folder, _ := req.GetFolderID()
			switch folder {
			case 0:
				ch := &tg.Channel{ID: 1, Title: "M", Broadcast: true}
				ch.SetAccessHash(1)
				return &tg.MessagesDialogs{
					Dialogs:  []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10}},
					Messages: []tg.MessageClass{&tg.Message{ID: 10, Date: 1700000010}},
					Chats:    []tg.ChatClass{ch},
				}, nil
			case 1:
				// Two archived channels — but limit will only take one.
				ch1 := &tg.Channel{ID: 2, Title: "A1", Broadcast: true}
				ch1.SetAccessHash(2)
				ch2 := &tg.Channel{ID: 3, Title: "A2", Broadcast: true}
				ch2.SetAccessHash(3)
				d1 := &tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}, TopMessage: 5}
				d1.SetFolderID(1)
				d2 := &tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 3}, TopMessage: 4}
				d2.SetFolderID(1)
				return &tg.MessagesDialogs{
					Dialogs:  []tg.DialogClass{d1, d2},
					Messages: []tg.MessageClass{&tg.Message{ID: 5, Date: 1700000005}, &tg.Message{ID: 4, Date: 1700000004}},
					Chats:    []tg.ChatClass{ch1, ch2},
				}, nil
			}
			return &tg.MessagesDialogs{}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)

	got, err := s.List(context.Background(), ListRequest{Archived: true, Limit: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got.Returned != 2 {
		t.Fatalf("Returned = %d, want 2 (1 main + 1 archived)", got.Returned)
	}
	if got.Cursor == "" {
		t.Fatal("expected non-empty cursor (one archived dialog was discarded)")
	}
	// Cursor must point at folder 1 (archive), at the last item we kept
	// (the first archived channel, A1 with id=2).
	c, err := DecodeCursor(got.Cursor)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if c.Folder != 1 {
		t.Errorf("cursor.Folder = %d, want 1 (archive)", c.Folder)
	}
	if c.OffsetPeerType != "channel" || c.OffsetPeerID != 2 {
		t.Errorf("cursor points at peer (%s,%d), want (channel,2)", c.OffsetPeerType, c.OffsetPeerID)
	}
	if c.OffsetID != 5 {
		t.Errorf("cursor.OffsetID = %d, want 5 (last in-limit message id)", c.OffsetID)
	}
	_ = calls
}

// TestService_List_LimitMidPageEmitsCursorFromKeptItem covers the same
// principle within a single folder: --limit cuts a page in half, so the
// emitted cursor must be derived from the last kept item — NOT from the
// Telegram-supplied next-page cursor that points past the dropped tail.
func TestService_List_LimitMidPageEmitsCursorFromKeptItem(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch1 := &tg.Channel{ID: 1, Title: "C1", Broadcast: true}
			ch1.SetAccessHash(11)
			ch2 := &tg.Channel{ID: 2, Title: "C2", Broadcast: true}
			ch2.SetAccessHash(22)
			ch3 := &tg.Channel{ID: 3, Title: "C3", Broadcast: true}
			ch3.SetAccessHash(33)
			// Slice form, Count > len(dialogs) → hasMore=true → would normally
			// produce a Telegram-supplied next cursor pointing past id=3.
			return &tg.MessagesDialogsSlice{
				Count: 100,
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 30},
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}, TopMessage: 20},
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 3}, TopMessage: 10},
				},
				Messages: []tg.MessageClass{
					&tg.Message{ID: 30, Date: 1700000030},
					&tg.Message{ID: 20, Date: 1700000020},
					&tg.Message{ID: 10, Date: 1700000010},
				},
				Chats: []tg.ChatClass{ch1, ch2, ch3},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)

	got, err := s.List(context.Background(), ListRequest{Limit: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got.Returned != 2 {
		t.Fatalf("Returned = %d, want 2", got.Returned)
	}
	if got.Cursor == "" {
		t.Fatal("expected non-empty cursor")
	}
	c, err := DecodeCursor(got.Cursor)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if c.OffsetID != 20 {
		t.Errorf("cursor.OffsetID = %d, want 20 (last in-limit msg id, not 10 which is past the limit)", c.OffsetID)
	}
	if c.OffsetPeerID != 2 {
		t.Errorf("cursor.OffsetPeerID = %d, want 2 (last in-limit peer)", c.OffsetPeerID)
	}
}

// TestService_List_ReturnedEqualsSourcesLen verifies that the Returned field
// in ListResult is the post-filter, post-limit length — and specifically that
// applying --type yields Returned < Total.
func TestService_List_ReturnedEqualsSourcesLen(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "C", Broadcast: true}
			ch.SetAccessHash(1)
			user := &tg.User{ID: 2}
			user.SetFirstName("Bob")
			user.SetAccessHash(2)
			return &tg.MessagesDialogs{
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10},
					&tg.Dialog{Peer: &tg.PeerUser{UserID: 2}, TopMessage: 20},
				},
				Messages: []tg.MessageClass{
					&tg.Message{ID: 10, Date: 1700000000},
					&tg.Message{ID: 20, Date: 1700000001},
				},
				Chats: []tg.ChatClass{ch},
				Users: []tg.UserClass{user},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)

	got, err := s.List(context.Background(), ListRequest{Types: []string{"user"}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got.Returned != len(got.Sources) {
		t.Errorf("Returned = %d, want %d (= len sources)", got.Returned, len(got.Sources))
	}
	if got.Returned != 1 {
		t.Errorf("Returned = %d, want 1 (only the user)", got.Returned)
	}
	if got.Total != 2 {
		t.Errorf("Total = %d, want 2 (unfiltered)", got.Total)
	}
}

// TestService_Inspect_NumericIDCacheMissReturnsError verifies that inspecting
// by `id:<n>` for a channel/user not in the peer cache fails with a clear
// error message instead of silently returning a peer with AccessHash=0 (which
// makes Telegram return garbage with empty fields).
func TestService_Inspect_NumericIDCacheMissReturnsError(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "peers.db")
	cache, err := resolve.NewPeerCache(cachePath)
	if err != nil {
		t.Fatalf("NewPeerCache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	resolver := resolve.NewResolver(nil, cache)
	s := New(&mockAPI{}, resolver, nil, nil, 0)

	// Positive ID (user) — no cache entry → must error out.
	_, err = s.Inspect(context.Background(), InspectRequest{Ref: "id:489000", NoStats: true})
	if err == nil {
		t.Fatal("expected error for cache-miss numeric user ID, got nil")
	}
	if !strings.Contains(err.Error(), "access hash") {
		t.Errorf("error = %q, want mention of 'access hash'", err.Error())
	}

	// Negative channel ID — same expectation.
	_, err = s.Inspect(context.Background(), InspectRequest{Ref: "id:-1009999999999", NoStats: true})
	if err == nil {
		t.Fatal("expected error for cache-miss numeric channel ID, got nil")
	}
}

func TestTelegramErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"channel private with arg", errors.New("messages.search: rpc error code 400: CHANNEL_PRIVATE (...)"), "CHANNEL_PRIVATE"},
		{"no arg", errors.New("rpc error code 400: USERNAME_INVALID"), "USERNAME_INVALID"},
		{"wrapped chain", fmt.Errorf("messages.getDialogs: %w", errors.New("rpc error code 400: PEER_ID_INVALID (caused by msg)")), "PEER_ID_INVALID"},
		{"non-uppercase fallback", errors.New("totally unexpected: lowercase"), "ERROR"},
		{"plain error fallback", errors.New("io: timeout"), "ERROR"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := telegramErrorCode(tc.err); got != tc.want {
				t.Errorf("telegramErrorCode(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}
