package sources

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

// mockAPI implements the sources.API interface, with the relevant methods set
// per test. Unused methods panic.
type mockAPI struct {
	getDialogs       func(ctx context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error)
	search           func(ctx context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error)
	getHistory       func(ctx context.Context, req *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error)
	getFullChannel   func(ctx context.Context, ch tg.InputChannelClass) (*tg.MessagesChatFull, error)
	getFullChat      func(ctx context.Context, chatID int64) (*tg.MessagesChatFull, error)
	getFullUser      func(ctx context.Context, u tg.InputUserClass) (*tg.UsersUserFull, error)
	getDialogFilters func(ctx context.Context) (*tg.MessagesDialogFilters, error)
}

func (m *mockAPI) MessagesGetDialogs(ctx context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
	return m.getDialogs(ctx, req)
}
func (m *mockAPI) MessagesSearch(ctx context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
	return m.search(ctx, req)
}
func (m *mockAPI) MessagesGetHistory(ctx context.Context, req *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
	return m.getHistory(ctx, req)
}
func (m *mockAPI) ChannelsGetFullChannel(ctx context.Context, ch tg.InputChannelClass) (*tg.MessagesChatFull, error) {
	return m.getFullChannel(ctx, ch)
}
func (m *mockAPI) MessagesGetFullChat(ctx context.Context, chatID int64) (*tg.MessagesChatFull, error) {
	return m.getFullChat(ctx, chatID)
}
func (m *mockAPI) UsersGetFullUser(ctx context.Context, u tg.InputUserClass) (*tg.UsersUserFull, error) {
	return m.getFullUser(ctx, u)
}
func (m *mockAPI) MessagesGetDialogFilters(ctx context.Context) (*tg.MessagesDialogFilters, error) {
	return m.getDialogFilters(ctx)
}

func TestFetchDialogs_SinglePage(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 111, Title: "Ch", Broadcast: true}
			ch.SetAccessHash(7)
			ch.SetUsername("ch")
			return &tg.MessagesDialogs{
				Dialogs: []tg.DialogClass{
					&tg.Dialog{
						Peer:        &tg.PeerChannel{ChannelID: 111},
						TopMessage:  500,
						UnreadCount: 3,
					},
				},
				Messages: []tg.MessageClass{
					&tg.Message{ID: 500, Date: 1700000000},
				},
				Chats: []tg.ChatClass{ch},
				Users: []tg.UserClass{},
			}, nil
		},
	}
	s := &Service{api: api}
	items, nextCursor, total, err := s.fetchDialogs(context.Background(), nil, 100, 0, 0)
	if err != nil {
		t.Fatalf("fetchDialogs: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	src := items[0].Source
	if src.ID != -1000000000111 {
		t.Errorf("ID = %d", src.ID)
	}
	if src.UnreadCount != 3 {
		t.Errorf("UnreadCount = %d", src.UnreadCount)
	}
	if src.LastMessage == nil || src.LastMessage.ID != 500 {
		t.Errorf("LastMessage = %+v", src.LastMessage)
	}
	if ch, ok := items[0].Peer.(*tg.InputPeerChannel); !ok || ch.AccessHash != 7 {
		t.Errorf("Peer = %+v, want InputPeerChannel with access hash 7", items[0].Peer)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if nextCursor != nil {
		t.Errorf("nextCursor should be nil for non-slice response, got %+v", nextCursor)
	}
}

func TestFetchDialogs_SliceWithCursor(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 222, Title: "C2", Broadcast: true}
			ch.SetAccessHash(9)
			return &tg.MessagesDialogsSlice{
				Count: 500,
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 222}, TopMessage: 42},
				},
				Messages: []tg.MessageClass{&tg.Message{ID: 42, Date: 1700000000}},
				Chats:    []tg.ChatClass{ch},
				Users:    []tg.UserClass{},
			}, nil
		},
	}
	s := &Service{api: api}
	items, nextCursor, total, err := s.fetchDialogs(context.Background(), nil, 1, 0, 0)
	if err != nil {
		t.Fatalf("fetchDialogs: %v", err)
	}
	if len(items) != 1 || total != 500 {
		t.Errorf("len=%d total=%d", len(items), total)
	}
	if nextCursor == nil {
		t.Fatal("expected non-nil cursor when more pages remain")
	}
	if nextCursor.OffsetID != 42 || nextCursor.OffsetPeerType != "channel" || nextCursor.OffsetPeerID != 222 {
		t.Errorf("cursor = %+v", nextCursor)
	}
}
