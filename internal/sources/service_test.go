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
