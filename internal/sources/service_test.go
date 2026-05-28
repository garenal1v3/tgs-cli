package sources

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gotd/td/tg"
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
	defer cache.Close()
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
	defer cache.Close()
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

func TestService_Inspect_NotSubscribedSkipsAllCalls(t *testing.T) {
	called := false
	api := &mockAPI{
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			called = true
			return nil, errors.New("should not be called")
		},
	}
	s := New(api, nil, nil, nil, 0)
	notSubscribed := false
	src, err := s.inspectKnownWithSubscription(context.Background(), Source{
		ID: -1000000000001, Type: "channel",
	}, &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1}, false, &notSubscribed)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if called {
		t.Error("fetchFull should be skipped when subscribed is *false")
	}
	if src.Subscribed == nil || *src.Subscribed != false {
		t.Errorf("Subscribed = %+v, want *false", src.Subscribed)
	}
	if src.Stats != nil {
		t.Errorf("Stats should be nil, got %+v", src.Stats)
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
