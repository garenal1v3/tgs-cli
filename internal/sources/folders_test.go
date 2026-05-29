package sources

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gotd/td/tg"
)

func TestService_Folders_SkipsDefaultIncludesCustomAndChatlist(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "A", Broadcast: true}
			ch.SetAccessHash(1)
			return &tg.MessagesDialogs{
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 1},
				},
				Messages: []tg.MessageClass{&tg.Message{ID: 1, Date: 1700000000}},
				Chats:    []tg.ChatClass{ch},
			}, nil
		},
		getDialogFilters: func(_ context.Context) (*tg.MessagesDialogFilters, error) {
			return &tg.MessagesDialogFilters{
				Filters: []tg.DialogFilterClass{
					&tg.DialogFilterDefault{},
					&tg.DialogFilter{
						ID:    2,
						Title: tg.TextWithEntities{Text: "Custom"},
						IncludePeers: []tg.InputPeerClass{
							&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
						},
					},
					&tg.DialogFilterChatlist{
						ID:    5,
						Title: tg.TextWithEntities{Text: "Shared"},
						IncludePeers: []tg.InputPeerClass{
							&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
						},
					},
				},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, err := s.Folders(context.Background())
	if err != nil {
		t.Fatalf("Folders: %v", err)
	}
	if got.Total != 2 || len(got.Folders) != 2 {
		t.Fatalf("Total=%d Folders=%d, want 2/2 (default skipped)", got.Total, len(got.Folders))
	}
	if got.Folders[0].Kind != "custom" || got.Folders[0].ID != 2 {
		t.Errorf("folder[0] = %+v, want kind=custom id=2", got.Folders[0])
	}
	if got.Folders[1].Kind != "chatlist" || got.Folders[1].ID != 5 {
		t.Errorf("folder[1] = %+v, want kind=chatlist id=5", got.Folders[1])
	}
	if len(got.Folders[0].Chats) != 1 || len(got.Folders[1].Chats) != 1 {
		t.Errorf("each folder should have 1 chat (the shared channel)")
	}
	if got.Folders[0].ChatsCount != 1 {
		t.Errorf("ChatsCount = %d, want 1", got.Folders[0].ChatsCount)
	}
}

func TestService_Folders_EmptyWhenNoCustom(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			return &tg.MessagesDialogs{}, nil
		},
		getDialogFilters: func(_ context.Context) (*tg.MessagesDialogFilters, error) {
			return &tg.MessagesDialogFilters{
				Filters: []tg.DialogFilterClass{&tg.DialogFilterDefault{}},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, err := s.Folders(context.Background())
	if err != nil {
		t.Fatalf("Folders: %v", err)
	}
	if got.Total != 0 || len(got.Folders) != 0 {
		t.Errorf("got Total=%d Folders=%d, want 0/0", got.Total, len(got.Folders))
	}
}

// A folder that resolves to zero chats must serialise "chats": [] (never
// null), so consumers can iterate the array unconditionally.
func TestService_Folders_EmptyFolderChatsIsEmptyArrayNotNull(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "Unrelated", Broadcast: true}
			ch.SetAccessHash(1)
			return &tg.MessagesDialogs{
				Dialogs:  []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 1}},
				Messages: []tg.MessageClass{&tg.Message{ID: 1, Date: 1700000000}},
				Chats:    []tg.ChatClass{ch},
			}, nil
		},
		getDialogFilters: func(_ context.Context) (*tg.MessagesDialogFilters, error) {
			return &tg.MessagesDialogFilters{
				Filters: []tg.DialogFilterClass{
					// Custom folder with no pinned/include peers and no auto-flags.
					&tg.DialogFilter{ID: 2, Title: tg.TextWithEntities{Text: "Empty"}},
				},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, err := s.Folders(context.Background())
	if err != nil {
		t.Fatalf("Folders: %v", err)
	}
	if len(got.Folders) != 1 {
		t.Fatalf("want 1 folder, got %d", len(got.Folders))
	}
	if got.Folders[0].Chats == nil {
		t.Error("Chats is nil; want non-nil empty slice")
	}
	b, _ := json.Marshal(got.Folders[0])
	if !strings.Contains(string(b), `"chats":[]`) {
		t.Errorf("JSON should contain \"chats\":[], got %s", b)
	}
}

func TestService_Folders_SameChatInTwoFolders_IndependentPinned(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "Shared", Broadcast: true}
			ch.SetAccessHash(1)
			return &tg.MessagesDialogs{
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 1},
				},
				Messages: []tg.MessageClass{&tg.Message{ID: 1, Date: 1700000000}},
				Chats:    []tg.ChatClass{ch},
			}, nil
		},
		getDialogFilters: func(_ context.Context) (*tg.MessagesDialogFilters, error) {
			return &tg.MessagesDialogFilters{
				Filters: []tg.DialogFilterClass{
					&tg.DialogFilter{
						ID:    2,
						Title: tg.TextWithEntities{Text: "A"},
						PinnedPeers: []tg.InputPeerClass{
							&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
						},
					},
					&tg.DialogFilter{
						ID:    3,
						Title: tg.TextWithEntities{Text: "B"},
						IncludePeers: []tg.InputPeerClass{
							&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
						},
					},
				},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, _ := s.Folders(context.Background())
	if !got.Folders[0].Chats[0].Pinned {
		t.Errorf("folder A: chat should be Pinned (came from pinned_peers)")
	}
	if got.Folders[1].Chats[0].Pinned {
		t.Errorf("folder B: chat should NOT be Pinned (came from include_peers)")
	}
}
