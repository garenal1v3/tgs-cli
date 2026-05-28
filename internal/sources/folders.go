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
// req.Archived = true includes archived dialogs in the source pool, so folders
// without `exclude_archived` see them; the filter's own `exclude_archived`
// flag still applies on top.
func (s *Service) Folders(ctx context.Context, req FoldersRequest) (*FoldersResult, error) {
	filters, err := s.fetchDialogFilters(ctx)
	if err != nil {
		return nil, err
	}
	dialogs, err := s.fetchAllDialogs(ctx, req.Archived)
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
		chats := matchFolder(f, dialogs, selfID)
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
		chats := matchFolder(f, dialogs, selfID)
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

// fetchAllDialogs walks all pages of getDialogs for folder 0 (and folder 1 if
// archived) — mirroring the List() walk but unconditionally exhaustive.
// The returned slice is the union of both folders' dialogs.
func (s *Service) fetchAllDialogs(ctx context.Context, archived bool) ([]sourceWithPeer, error) {
	const pageSize = 100
	var out []sourceWithPeer

	folders := []int{0}
	if archived {
		folders = append(folders, 1)
	}
	for _, folderID := range folders {
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
