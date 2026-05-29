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
	got, err := s.ResolveFolder(context.Background(), "2")
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
	got, err := s.ResolveFolder(context.Background(), "crypto")
	if err != nil {
		t.Fatalf("ResolveFolder: %v", err)
	}
	if got.Folder.ID != 2 {
		t.Errorf("ID = %d, want 2", got.Folder.ID)
	}
}

func TestResolveFolder_NotFound(t *testing.T) {
	s := New(threeFilterAPI(), nil, nil, nil, 0)
	_, err := s.ResolveFolder(context.Background(), "Nope")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "folder not found") {
		t.Errorf("err = %q, want 'folder not found'", err)
	}
}

func TestResolveFolder_DefaultIsNotResolvable(t *testing.T) {
	s := New(threeFilterAPI(), nil, nil, nil, 0)
	if _, err := s.ResolveFolder(context.Background(), "0"); err == nil {
		t.Errorf("expected not-found for id=0 (default)")
	}
	if _, err := s.ResolveFolder(context.Background(), "All chats"); err == nil {
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
	_, err := s.ResolveFolder(context.Background(), "dup")
	if err == nil {
		t.Fatal("expected ambiguous-name error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("err = %q, want 'ambiguous'", err)
	}
}

// ResolveFolder must walk BOTH the main folder (0) and the archive (1) so a
// folder whose members are archived still resolves to its full contents, with
// Folder.Chats and Peers kept strictly parallel.
func TestResolveFolder_IncludesArchivedChats(t *testing.T) {
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
		getDialogFilters: func(_ context.Context) (*tg.MessagesDialogFilters, error) {
			return &tg.MessagesDialogFilters{
				Filters: []tg.DialogFilterClass{
					&tg.DialogFilter{
						ID:    2,
						Title: tg.TextWithEntities{Text: "Mixed"},
						IncludePeers: []tg.InputPeerClass{
							&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
							&tg.InputPeerChannel{ChannelID: 2, AccessHash: 2},
						},
					},
				},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)
	got, err := s.ResolveFolder(context.Background(), "Mixed")
	if err != nil {
		t.Fatalf("ResolveFolder: %v", err)
	}
	if len(foldersSeen) != 2 || foldersSeen[0] != 0 || foldersSeen[1] != 1 {
		t.Errorf("foldersSeen = %v, want [0 1] (archive must be walked)", foldersSeen)
	}
	if len(got.Folder.Chats) != 2 {
		t.Fatalf("Chats = %d, want 2 (main + archived)", len(got.Folder.Chats))
	}
	if len(got.Peers) != len(got.Folder.Chats) {
		t.Errorf("Peers (%d) and Chats (%d) must be parallel", len(got.Peers), len(got.Folder.Chats))
	}
	var archivedFound bool
	for _, c := range got.Folder.Chats {
		if c.Archived {
			archivedFound = true
		}
	}
	if !archivedFound {
		t.Error("expected the archived channel to be present in folder contents")
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
	got, err := s.ResolveFolder(context.Background(), "shared community")
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
