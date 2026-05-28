package sources

import (
	"context"
	"strings"
	"testing"

	"github.com/gotd/td/tg"
)

func threeFilterAPI() *mockAPI {
	return &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "Ch", Broadcast: true}
			ch.SetAccessHash(7)
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
						Title: tg.TextWithEntities{Text: "Crypto"},
						IncludePeers: []tg.InputPeerClass{
							&tg.InputPeerChannel{ChannelID: 1, AccessHash: 7},
						},
					},
					&tg.DialogFilter{
						ID:    3,
						Title: tg.TextWithEntities{Text: "Work"},
					},
				},
			}, nil
		},
	}
}

func TestResolveFolder_ByID(t *testing.T) {
	s := New(threeFilterAPI(), nil, nil, nil, 0)
	got, err := s.ResolveFolder(context.Background(), "2", false)
	if err != nil {
		t.Fatalf("ResolveFolder: %v", err)
	}
	if got.Folder.ID != 2 || got.Folder.Title != "Crypto" {
		t.Errorf("Folder = %+v", got.Folder)
	}
	if len(got.Peers) != 1 {
		t.Errorf("len(Peers) = %d, want 1", len(got.Peers))
	}
}

func TestResolveFolder_ByName_CaseInsensitive(t *testing.T) {
	s := New(threeFilterAPI(), nil, nil, nil, 0)
	got, err := s.ResolveFolder(context.Background(), "crypto", false)
	if err != nil {
		t.Fatalf("ResolveFolder: %v", err)
	}
	if got.Folder.ID != 2 {
		t.Errorf("ID = %d, want 2", got.Folder.ID)
	}
}

func TestResolveFolder_NotFound(t *testing.T) {
	s := New(threeFilterAPI(), nil, nil, nil, 0)
	_, err := s.ResolveFolder(context.Background(), "Nope", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "folder not found") {
		t.Errorf("err = %q, want 'folder not found'", err)
	}
}

func TestResolveFolder_DefaultIsNotResolvable(t *testing.T) {
	s := New(threeFilterAPI(), nil, nil, nil, 0)
	if _, err := s.ResolveFolder(context.Background(), "0", false); err == nil {
		t.Errorf("expected not-found for id=0 (default)")
	}
	if _, err := s.ResolveFolder(context.Background(), "All chats", false); err == nil {
		t.Errorf("expected not-found for name 'All chats'")
	}
}

func TestResolveFolder_AmbiguousName(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			return &tg.MessagesDialogs{}, nil
		},
		getDialogFilters: func(_ context.Context) (*tg.MessagesDialogFilters, error) {
			return &tg.MessagesDialogFilters{
				Filters: []tg.DialogFilterClass{
					&tg.DialogFilter{ID: 2, Title: tg.TextWithEntities{Text: "Dup"}},
					&tg.DialogFilter{ID: 5, Title: tg.TextWithEntities{Text: "DUP"}},
				},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	_, err := s.ResolveFolder(context.Background(), "dup", false)
	if err == nil {
		t.Fatal("expected ambiguous-name error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("err = %q, want 'ambiguous'", err)
	}
}

func TestResolveFolder_ByChatlistName(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			ch := &tg.Channel{ID: 1, Title: "Ch", Broadcast: true}
			ch.SetAccessHash(7)
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
					&tg.DialogFilterChatlist{
						ID:    7,
						Title: tg.TextWithEntities{Text: "Shared Community"},
						IncludePeers: []tg.InputPeerClass{
							&tg.InputPeerChannel{ChannelID: 1, AccessHash: 7},
						},
					},
				},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, err := s.ResolveFolder(context.Background(), "shared community", false)
	if err != nil {
		t.Fatalf("ResolveFolder: %v", err)
	}
	if got.Folder.ID != 7 || got.Folder.Kind != "chatlist" {
		t.Errorf("Folder = %+v, want id=7 kind=chatlist", got.Folder)
	}
	if len(got.Peers) != 1 {
		t.Errorf("len(Peers) = %d, want 1", len(got.Peers))
	}
}
