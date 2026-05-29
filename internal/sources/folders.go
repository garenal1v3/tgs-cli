package sources

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"

	"github.com/searchtgcli/tgs/internal/retry"
)

// Folders enumerates the user's dialog filters and returns each filter
// (except DialogFilterDefault) with its contents resolved against the user's
// dialog list.
//
// A folder is itself a view that can contain archived chats, so the source
// pool ALWAYS includes archived dialogs (folder 1) in addition to the main
// folder (0) — mirroring what the Telegram client shows when you open a
// folder. The filter's own `exclude_archived` flag still applies on top.
func (s *Service) Folders(ctx context.Context) (*FoldersResult, error) {
	filters, err := s.fetchDialogFilters(ctx)
	if err != nil {
		return nil, err
	}
	dialogs, err := s.fetchAllDialogs(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Folder, 0, len(filters))
	for _, filter := range filters {
		f, skip := buildFolder(filter, dialogs, s.selfID)
		if skip {
			continue
		}
		out = append(out, f)
	}

	return &FoldersResult{
		Folders: out,
		Total:   len(out),
	}, nil
}

// buildFolder turns one DialogFilterClass into a Folder by classifying its
// kind and running matchFolder. Returns skip=true for DialogFilterDefault.
func buildFolder(filter tg.DialogFilterClass, dialogs []sourceWithPeer, selfID int64) (Folder, bool) {
	switch f := filter.(type) {
	case *tg.DialogFilterDefault:
		return Folder{}, true
	case *tg.DialogFilter:
		chats := matchFolderChats(f, dialogs, selfID)
		emoticon, _ := f.GetEmoticon()
		return Folder{
			ID:         f.ID,
			Kind:       "custom",
			Title:      f.Title.Text,
			Emoticon:   emoticon,
			ChatsCount: len(chats),
			Chats:      chats,
		}, false
	case *tg.DialogFilterChatlist:
		chats := matchFolderChats(f, dialogs, selfID)
		emoticon, _ := f.GetEmoticon()
		return Folder{
			ID:           f.ID,
			Kind:         "chatlist",
			Title:        f.Title.Text,
			Emoticon:     emoticon,
			HasMyInvites: f.HasMyInvites,
			ChatsCount:   len(chats),
			Chats:        chats,
		}, false
	}
	return Folder{}, true
}

// matchFolderChats wraps matchFolder and guarantees a non-nil slice so the
// JSON `chats` field is always an array (`[]`), never `null` — empty folders
// would otherwise break consumers that iterate the array unconditionally.
func matchFolderChats(filter tg.DialogFilterClass, dialogs []sourceWithPeer, selfID int64) []Source {
	chats := matchFolder(filter, dialogs, selfID)
	if chats == nil {
		return []Source{}
	}
	return chats
}

// fetchDialogFilters wraps messages.getDialogFilters with retry.
//
// Note: gotd/td returns *tg.MessagesDialogFilters directly here (no polymorphic
// MessagesDialogFiltersClass interface is generated for this method, unlike
// MessagesGetDialogs which returns MessagesDialogsClass).
func (s *Service) fetchDialogFilters(ctx context.Context) ([]tg.DialogFilterClass, error) {
	res, err := retry.Do(ctx, s.retry, func() (*tg.MessagesDialogFilters, error) {
		r, err := s.api.MessagesGetDialogFilters(ctx)
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.getDialogFilters: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("messages.getDialogFilters: nil response")
	}
	return res.Filters, nil
}

// fetchAllDialogs walks all pages of getDialogs for both the main folder (0)
// and the archive folder (1) — mirroring the List() walk but unconditionally
// exhaustive. The returned slice is the union of both folders' dialogs.
//
// The archive is always included: a user-defined folder is a cross-cutting
// view that may contain archived chats, and the Telegram client shows them
// when you open the folder. Folder resolution must reproduce that, so callers
// don't get a misleadingly empty folder just because its chats are archived.
func (s *Service) fetchAllDialogs(ctx context.Context) ([]sourceWithPeer, error) {
	const pageSize = 100
	var out []sourceWithPeer

	for _, folderID := range []int{0, 1} {
		var cur *Cursor
		for {
			page, next, _, err := s.fetchDialogs(ctx, cur, pageSize, folderID, s.selfID)
			if err != nil {
				return nil, err
			}
			out = append(out, page...)
			if next == nil {
				break
			}
			cur = next
		}
	}
	return out, nil
}
