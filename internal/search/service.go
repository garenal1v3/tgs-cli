package search

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/gotd/td/tg"

	"github.com/searchtgcli/tgs/internal/resolve"
	"github.com/searchtgcli/tgs/internal/retry"
)

// SearchRequest describes a search within one or more specific chats.
type SearchRequest struct {
	Peers   []tg.InputPeerClass
	Query   string
	FromID  tg.InputPeerClass
	Filter  tg.MessagesFilterClass
	After   int
	Before  int
	TopicID int
	Limit   int
	Cursor  string
}

// GlobalSearchRequest describes a cross-chat global search.
type GlobalSearchRequest struct {
	Query        string
	Filter       tg.MessagesFilterClass
	After        int
	Before       int
	Archived     bool
	ChannelsOnly bool
	GroupsOnly   bool
	UsersOnly    bool
	Limit        int
	Cursor       string
}

// Service wraps Telegram search API calls with retry, pagination,
// and message conversion.
type Service struct {
	api      *tg.Client
	resolver *resolve.Resolver
	retry    *retry.Policy
}

// NewService creates a Service. If retryPolicy is nil, DefaultPolicy is used.
func NewService(api *tg.Client, resolver *resolve.Resolver, retryPolicy *retry.Policy) *Service {
	if retryPolicy == nil {
		retryPolicy = retry.DefaultPolicy()
	}
	return &Service{
		api:      api,
		resolver: resolver,
		retry:    retryPolicy,
	}
}

// Search searches within one or more chats.
func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	if len(req.Peers) == 0 {
		return nil, errors.New("at least one peer is required")
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Filter == nil {
		req.Filter = &tg.InputMessagesFilterEmpty{}
	}

	if len(req.Peers) == 1 {
		return s.searchSingle(ctx, req.Peers[0], req)
	}
	return s.searchMulti(ctx, req)
}

// searchSingle performs a search within a single peer.
func (s *Service) searchSingle(ctx context.Context, peer tg.InputPeerClass, req SearchRequest) (*SearchResult, error) {
	cur, err := DecodeCursor(req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}

	apiReq := &tg.MessagesSearchRequest{
		Peer:    peer,
		Q:       req.Query,
		Filter:  req.Filter,
		MinDate: req.After,
		MaxDate: req.Before,
		Limit:   req.Limit,
	}

	if req.FromID != nil {
		apiReq.SetFromID(req.FromID)
	}
	if req.TopicID != 0 {
		apiReq.SetTopMsgID(req.TopicID)
	}
	if cur != nil {
		apiReq.OffsetID = cur.OffsetID
	}

	res, err := retry.Do(ctx, s.retry, func() (tg.MessagesMessagesClass, error) {
		r, err := s.api.MessagesSearch(ctx, apiReq)
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.search: %w", err)
	}

	return s.convertMessages(res, req.Limit)
}

// searchMulti fans out to each peer sequentially and merges results.
func (s *Service) searchMulti(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	mc, err := DecodeMultiCursor(req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("decode multi-cursor: %w", err)
	}

	// Build cursor map, initialise if absent.
	if mc == nil {
		mc = &MultiCursor{
			Chats: make([]ChatCursor, len(req.Peers)),
		}
		for i, peer := range req.Peers {
			mc.Chats[i] = ChatCursor{PeerID: peerID(peer)}
		}
	}

	cursorMap := make(map[int64]*ChatCursor, len(mc.Chats))
	for i := range mc.Chats {
		cursorMap[mc.Chats[i].PeerID] = &mc.Chats[i]
	}

	var allMessages []Message
	total := 0

	for _, peer := range req.Peers {
		pid := peerID(peer)
		cc := cursorMap[pid]
		if cc != nil && cc.Done {
			continue
		}

		singleReq := req
		singleReq.Peers = nil
		if cc != nil && cc.OffsetID != 0 {
			singleReq.Cursor = (&Cursor{OffsetID: cc.OffsetID}).Encode()
		} else {
			singleReq.Cursor = ""
		}

		result, err := s.searchSingle(ctx, peer, singleReq)
		if err != nil {
			return nil, fmt.Errorf("search peer %d: %w", pid, err)
		}

		allMessages = append(allMessages, result.Messages...)
		total += result.Total

		// Update cursor state for this peer.
		if cc != nil {
			if result.Cursor == "" {
				cc.Done = true
			} else {
				decoded, err := DecodeCursor(result.Cursor)
				if err == nil && decoded != nil {
					cc.OffsetID = decoded.OffsetID
				}
			}
		}
	}

	// Sort merged messages by date descending.
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].Date > allMessages[j].Date
	})

	// Truncate to limit.
	if len(allMessages) > req.Limit {
		allMessages = allMessages[:req.Limit]
	}

	result := &SearchResult{
		Messages: allMessages,
		Total:    total,
	}

	// Encode multi-cursor if not all peers are done.
	allDone := true
	for _, cc := range mc.Chats {
		if !cc.Done {
			allDone = false
			break
		}
	}
	if !allDone {
		result.Cursor = mc.Encode()
	}

	return result, nil
}

// SearchGlobal performs a global message search across all chats.
func (s *Service) SearchGlobal(ctx context.Context, req GlobalSearchRequest) (*SearchResult, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Filter == nil {
		req.Filter = &tg.InputMessagesFilterEmpty{}
	}

	cur, err := DecodeCursor(req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}

	apiReq := &tg.MessagesSearchGlobalRequest{
		Q:          req.Query,
		Filter:     req.Filter,
		MinDate:    req.After,
		MaxDate:    req.Before,
		Limit:      req.Limit,
		OffsetPeer: &tg.InputPeerEmpty{},
	}

	if req.ChannelsOnly {
		apiReq.SetBroadcastsOnly(true)
	}
	if req.GroupsOnly {
		apiReq.SetGroupsOnly(true)
	}
	if req.UsersOnly {
		apiReq.SetUsersOnly(true)
	}
	if req.Archived {
		apiReq.SetFolderID(1)
	}

	if cur != nil {
		apiReq.OffsetID = cur.OffsetID
		apiReq.OffsetRate = cur.OffsetRate
	}

	res, err := retry.Do(ctx, s.retry, func() (tg.MessagesMessagesClass, error) {
		r, err := s.api.MessagesSearchGlobal(ctx, apiReq)
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.searchGlobal: %w", err)
	}

	return s.convertMessages(res, req.Limit)
}

// convertMessages converts a Telegram MessagesMessagesClass response into
// a SearchResult with cursor.
func (s *Service) convertMessages(res tg.MessagesMessagesClass, limit int) (*SearchResult, error) {
	var (
		msgs     []tg.MessageClass
		chats    []tg.ChatClass
		users    []tg.UserClass
		total    int
		nextRate int
	)

	switch v := res.(type) {
	case *tg.MessagesMessages:
		msgs = v.Messages
		chats = v.Chats
		users = v.Users
		total = len(v.Messages)

	case *tg.MessagesMessagesSlice:
		msgs = v.Messages
		chats = v.Chats
		users = v.Users
		total = v.Count
		if rate, ok := v.GetNextRate(); ok {
			nextRate = rate
		}

	case *tg.MessagesChannelMessages:
		msgs = v.Messages
		chats = v.Chats
		users = v.Users
		total = v.Count

	default:
		return nil, fmt.Errorf("unexpected response type: %T", res)
	}

	chatMap := buildChatMap(chats)
	userMap := buildUserMap(users)

	messages := make([]Message, 0, len(msgs))
	for _, mc := range msgs {
		m, ok := mc.(*tg.Message)
		if !ok {
			continue // skip service messages, empty messages, etc.
		}
		messages = append(messages, convertMessage(m, chatMap, userMap))
	}

	result := &SearchResult{
		Messages: messages,
		Total:    total,
	}

	// Build cursor if there might be more results.
	if len(messages) >= limit && len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		cur := &Cursor{
			OffsetID: lastMsg.ID,
		}
		if nextRate != 0 {
			cur.OffsetRate = nextRate
		}
		result.Cursor = cur.Encode()
	}

	return result, nil
}

// convertMessage converts a single tg.Message into our Message model.
func convertMessage(m *tg.Message, chatMap map[int64]ChatInfo, userMap map[int64]UserInfo) Message {
	msg := Message{
		ID:   m.ID,
		Date: time.Unix(int64(m.Date), 0).UTC().Format(time.RFC3339),
		Text: m.Message,
	}

	// Chat info from PeerID. Lookups use the raw ID (chatMap key); emitted
	// ChatInfo.ID is always Bot-API form.
	if m.PeerID != nil {
		pid := peerClassID(m.PeerID)
		if ci, ok := chatMap[pid]; ok {
			msg.Chat = ci
		} else if _, isUser := m.PeerID.(*tg.PeerUser); isUser {
			ci := ChatInfo{ID: pid, Type: "private"} // user IDs are already Bot-API form
			if ui, ok := userMap[pid]; ok {
				ci.Title = userDisplayName(ui)
				ci.Username = ui.Username
			}
			msg.Chat = ci
		} else {
			msg.Chat = ChatInfo{ID: botAPIPeerID(m.PeerID)}
		}
	}

	// From info.
	if fromPeer, ok := m.GetFromID(); ok {
		fid := peerClassID(fromPeer)
		if ui, found := userMap[fid]; found {
			msg.From = &ui
		} else {
			msg.From = &UserInfo{ID: fid}
		}
	}

	// Media.
	if media, ok := m.GetMedia(); ok {
		msg.Media = extractMedia(media)
	}

	// Views & forwards.
	if views, ok := m.GetViews(); ok {
		msg.Views = views
	}
	if forwards, ok := m.GetForwards(); ok {
		msg.Forwards = forwards
	}

	// Replies count (channel post comments).
	if replies, ok := m.GetReplies(); ok {
		msg.Replies = replies.Replies
	}

	// Reactions.
	if reactions, ok := m.GetReactions(); ok {
		for _, r := range reactions.Results {
			rc := ReactionCount{Count: r.Count}
			switch v := r.Reaction.(type) {
			case *tg.ReactionEmoji:
				rc.Emoji = v.Emoticon
			case *tg.ReactionCustomEmoji:
				rc.Emoji = fmt.Sprintf("custom:%d", v.DocumentID)
			case *tg.ReactionPaid:
				rc.Emoji = "⭐"
			default:
				continue
			}
			msg.Reactions = append(msg.Reactions, rc)
		}
	}

	// Reply info.
	if replyTo, ok := m.GetReplyTo(); ok {
		if rh, ok := replyTo.(*tg.MessageReplyHeader); ok {
			if replyMsgID, ok := rh.GetReplyToMsgID(); ok {
				msg.ReplyToMsgID = replyMsgID
			}
			if topicID, ok := rh.GetReplyToTopID(); ok {
				msg.TopicID = topicID
			}
		}
	}

	return msg
}

// extractMedia extracts MediaInfo from a Telegram MessageMediaClass.
func extractMedia(media tg.MessageMediaClass) *MediaInfo {
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		info := &MediaInfo{Type: "photo"}
		if photo, ok := m.GetPhoto(); ok {
			if p, ok := photo.(*tg.Photo); ok {
				// Find the largest PhotoSize for width/height.
				for _, size := range p.Sizes {
					if ps, ok := size.(*tg.PhotoSize); ok {
						if ps.W > info.Width || ps.H > info.Height {
							info.Width = ps.W
							info.Height = ps.H
						}
					}
				}
			}
		}
		return info

	case *tg.MessageMediaDocument:
		if doc, ok := m.GetDocument(); ok {
			if d, ok := doc.(*tg.Document); ok {
				return extractDocumentMedia(d)
			}
		}
		return &MediaInfo{Type: "document"}

	case *tg.MessageMediaGeo:
		return &MediaInfo{Type: "geo"}

	case *tg.MessageMediaContact:
		return &MediaInfo{Type: "contact"}

	case *tg.MessageMediaWebPage:
		return &MediaInfo{Type: "webpage"}

	default:
		return nil
	}
}

// extractDocumentMedia determines the document subtype from its attributes.
func extractDocumentMedia(d *tg.Document) *MediaInfo {
	info := &MediaInfo{
		Type:     "document",
		FileSize: d.Size,
		MimeType: d.MimeType,
	}

	for _, attr := range d.Attributes {
		switch a := attr.(type) {
		case *tg.DocumentAttributeVideo:
			if a.RoundMessage {
				info.Type = "round_video"
			} else {
				info.Type = "video"
			}
			info.Duration = int(math.Round(a.Duration))
			info.Width = a.W
			info.Height = a.H

		case *tg.DocumentAttributeAudio:
			if a.Voice {
				info.Type = "voice"
			} else {
				info.Type = "audio"
			}
			info.Duration = a.Duration

		case *tg.DocumentAttributeAnimated:
			info.Type = "gif"

		case *tg.DocumentAttributeSticker:
			info.Type = "sticker"

		case *tg.DocumentAttributeFilename:
			info.FileName = a.FileName

		case *tg.DocumentAttributeImageSize:
			info.Width = a.W
			info.Height = a.H
		}
	}

	return info
}

// botAPIChannelID converts a raw Telegram channel ID to its Bot-API negative
// form (-100xxxxxxxxxx). Mirrors sources.botAPIChannelID so chat IDs are
// consistent across `sources` and `search` output and can be fed straight back
// into --chat / --folder.
func botAPIChannelID(id int64) int64 { return -1000000000000 - id }

// botAPIPeerID converts a PeerClass to a Bot-API ID: users stay positive,
// legacy groups become -id, channels/supergroups become -100…id.
func botAPIPeerID(peer tg.PeerClass) int64 {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return p.UserID
	case *tg.PeerChat:
		return -p.ChatID
	case *tg.PeerChannel:
		return botAPIChannelID(p.ChannelID)
	default:
		return 0
	}
}

// buildChatMap builds a lookup map from ChatClass slices. The map is keyed by
// the RAW Telegram ID (matching peerClassID lookups), while each ChatInfo.ID
// value is the Bot-API form for output consistency.
func buildChatMap(chats []tg.ChatClass) map[int64]ChatInfo {
	m := make(map[int64]ChatInfo, len(chats))
	for _, c := range chats {
		switch v := c.(type) {
		case *tg.Chat:
			m[v.ID] = ChatInfo{
				ID:    -v.ID,
				Type:  "group",
				Title: v.Title,
			}
		case *tg.Channel:
			ci := ChatInfo{
				ID:    botAPIChannelID(v.ID),
				Title: v.Title,
			}
			legacy, _ := v.GetUsername()
			extra, _ := v.GetUsernames()
			if un := activeUsername(legacy, extra); un != "" {
				ci.Username = un
			}
			if v.Broadcast {
				ci.Type = "channel"
			} else {
				ci.Type = "supergroup"
			}
			m[v.ID] = ci
		}
	}
	return m
}

// buildUserMap builds a lookup map from UserClass slices.
func buildUserMap(users []tg.UserClass) map[int64]UserInfo {
	m := make(map[int64]UserInfo, len(users))
	for _, u := range users {
		user, ok := u.(*tg.User)
		if !ok {
			continue
		}
		ui := UserInfo{ID: user.ID}
		if fn, ok := user.GetFirstName(); ok {
			ui.FirstName = fn
		}
		if ln, ok := user.GetLastName(); ok {
			ui.LastName = ln
		}
		legacy, _ := user.GetUsername()
		extra, _ := user.GetUsernames()
		if un := activeUsername(legacy, extra); un != "" {
			ui.Username = un
		}
		m[user.ID] = ui
	}
	return m
}

// activeUsername returns a peer's user-visible handle. Prefers the legacy
// Username field; if empty, falls back to the first Active entry in the
// collectible-usernames array (tg.Username with Active=true) — required for
// peers like @durov or @money whose primary handle now lives in that array.
// Without this fallback search results displayed such peers with an empty
// username and inferred private access.
func activeUsername(legacy string, list []tg.Username) string {
	if legacy != "" {
		return legacy
	}
	for _, u := range list {
		if u.Active && u.Username != "" {
			return u.Username
		}
	}
	return ""
}

// peerID extracts a numeric ID from an InputPeerClass.
func peerID(peer tg.InputPeerClass) int64 {
	switch p := peer.(type) {
	case *tg.InputPeerUser:
		return p.UserID
	case *tg.InputPeerChat:
		return p.ChatID
	case *tg.InputPeerChannel:
		return p.ChannelID
	default:
		return 0
	}
}

func userDisplayName(ui UserInfo) string {
	name := ui.FirstName
	if ui.LastName != "" {
		if name != "" {
			name += " "
		}
		name += ui.LastName
	}
	return name
}

// peerClassID extracts a numeric ID from a PeerClass.
func peerClassID(peer tg.PeerClass) int64 {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return p.UserID
	case *tg.PeerChat:
		return p.ChatID
	case *tg.PeerChannel:
		return p.ChannelID
	default:
		return 0
	}
}
