package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

// sourceWithPeer pairs a Source with the InputPeerClass needed to re-call
// Telegram on its behalf (used by --with-stats enrichment). Internal use only.
type sourceWithPeer struct {
	Source Source
	Peer   tg.InputPeerClass
}

// fetchDialogs makes one messages.getDialogs call and returns the converted
// items (Source + InputPeer for follow-up calls), a next-page cursor (nil if
// no more pages), and the server-reported total. archived selects folder 1
// (archive) when true, folder 0 otherwise. selfID is the authenticated user's
// ID, used for Saved Messages detection.
func (s *Service) fetchDialogs(ctx context.Context, cur *Cursor, limit int, archived bool, selfID int64) ([]sourceWithPeer, *Cursor, int, error) {
	req := &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      limit,
	}
	if archived {
		req.SetFolderID(1)
	} else {
		req.SetFolderID(0)
	}
	if cur != nil {
		req.OffsetDate = cur.OffsetDate
		req.OffsetID = cur.OffsetID
		req.OffsetPeer = cursorToInputPeer(cur)
	}

	res, err := s.api.MessagesGetDialogs(ctx, req)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("messages.getDialogs: %w", err)
	}

	var (
		dialogs []tg.DialogClass
		msgs    []tg.MessageClass
		chats   []tg.ChatClass
		users   []tg.UserClass
		total   int
		hasMore bool
	)
	switch v := res.(type) {
	case *tg.MessagesDialogs:
		dialogs, msgs, chats, users = v.Dialogs, v.Messages, v.Chats, v.Users
		total = len(dialogs)
	case *tg.MessagesDialogsSlice:
		dialogs, msgs, chats, users = v.Dialogs, v.Messages, v.Chats, v.Users
		total = v.Count
		hasMore = len(dialogs) >= limit && len(dialogs) > 0
	default:
		return nil, nil, 0, fmt.Errorf("unexpected dialogs response: %T", res)
	}

	chatByID := make(map[int64]tg.ChatClass, len(chats))
	for _, c := range chats {
		switch v := c.(type) {
		case *tg.Chat:
			chatByID[v.ID] = v
		case *tg.Channel:
			chatByID[v.ID] = v
		}
	}
	userByID := make(map[int64]*tg.User, len(users))
	for _, u := range users {
		if uu, ok := u.(*tg.User); ok {
			userByID[uu.ID] = uu
		}
	}
	msgByID := make(map[int]*tg.Message, len(msgs))
	for _, m := range msgs {
		if mm, ok := m.(*tg.Message); ok {
			msgByID[mm.ID] = mm
		}
	}

	out := make([]sourceWithPeer, 0, len(dialogs))
	var lastDate, lastID int
	var lastPeerType string
	var lastPeerID, lastPeerAccess int64

	for _, dc := range dialogs {
		d, ok := dc.(*tg.Dialog)
		if !ok {
			continue
		}
		src, peer, peerType, peerID, peerAccess := dialogToSource(d, chatByID, userByID, msgByID, selfID)
		if src.ID == 0 {
			continue // unknown peer type
		}
		out = append(out, sourceWithPeer{Source: src, Peer: peer})
		lastID = d.TopMessage
		if mm := msgByID[d.TopMessage]; mm != nil {
			lastDate = mm.Date
		}
		lastPeerType, lastPeerID, lastPeerAccess = peerType, peerID, peerAccess
	}

	if !hasMore {
		return out, nil, total, nil
	}
	next := &Cursor{
		OffsetDate:           lastDate,
		OffsetID:             lastID,
		OffsetPeerType:       lastPeerType,
		OffsetPeerID:         lastPeerID,
		OffsetPeerAccessHash: lastPeerAccess,
	}
	return out, next, total, nil
}

// dialogToSource enriches a chatToSource/userToSource result with dialog-level
// fields (unread, pinned, last_message). It also returns:
//   - the constructed InputPeer (with access hash) for follow-up Telegram calls
//   - the peer fingerprint (type/id/access_hash) needed for cursor construction
func dialogToSource(
	d *tg.Dialog,
	chatByID map[int64]tg.ChatClass,
	userByID map[int64]*tg.User,
	msgByID map[int]*tg.Message,
	selfID int64,
) (src Source, peer tg.InputPeerClass, peerType string, peerID, peerAccessHash int64) {
	switch p := d.Peer.(type) {
	case *tg.PeerChannel:
		c, ok := chatByID[p.ChannelID]
		if !ok {
			return Source{}, nil, "", 0, 0
		}
		ch := c.(*tg.Channel)
		src = chatToSource(ch)
		access, _ := ch.GetAccessHash()
		peer = &tg.InputPeerChannel{ChannelID: ch.ID, AccessHash: access}
		peerType, peerID, peerAccessHash = "channel", ch.ID, access
	case *tg.PeerChat:
		c, ok := chatByID[p.ChatID]
		if !ok {
			return Source{}, nil, "", 0, 0
		}
		src = chatToSource(c.(*tg.Chat))
		peer = &tg.InputPeerChat{ChatID: p.ChatID}
		peerType, peerID = "chat", p.ChatID
	case *tg.PeerUser:
		u, ok := userByID[p.UserID]
		if !ok {
			return Source{}, nil, "", 0, 0
		}
		src = userToSource(u, selfID)
		access, _ := u.GetAccessHash()
		peer = &tg.InputPeerUser{UserID: u.ID, AccessHash: access}
		peerType, peerID, peerAccessHash = "user", u.ID, access
	default:
		return Source{}, nil, "", 0, 0
	}

	src.UnreadCount = d.UnreadCount
	src.Pinned = d.Pinned
	if folderID, ok := d.GetFolderID(); ok && folderID == 1 {
		src.Archived = true
	}
	if mm := msgByID[d.TopMessage]; mm != nil {
		src.LastMessage = &LastMessage{
			ID:   mm.ID,
			Date: time.Unix(int64(mm.Date), 0).UTC().Format(time.RFC3339),
		}
	}
	return src, peer, peerType, peerID, peerAccessHash
}

// cursorToInputPeer reconstructs the offset_peer parameter for getDialogs.
func cursorToInputPeer(c *Cursor) tg.InputPeerClass {
	switch c.OffsetPeerType {
	case "user":
		return &tg.InputPeerUser{UserID: c.OffsetPeerID, AccessHash: c.OffsetPeerAccessHash}
	case "chat":
		return &tg.InputPeerChat{ChatID: c.OffsetPeerID}
	case "channel":
		return &tg.InputPeerChannel{ChannelID: c.OffsetPeerID, AccessHash: c.OffsetPeerAccessHash}
	default:
		return &tg.InputPeerEmpty{}
	}
}
