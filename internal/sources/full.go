package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"

	"github.com/garenal1v3/tgs-cli/internal/retry"
)

// fetchFull pulls per-source "full" info (members count, description, linked
// chat, invite link). It dispatches by Source.Type:
//   - "channel"/"supergroup" → channels.getFullChannel
//   - "group"                → messages.getFullChat
//   - "user"/"bot"           → users.getFullUser
//
// Returns a CachedPayload with Stats unset — Stats are filled separately by
// fetchStats (Task 7).
func (s *Service) fetchFull(ctx context.Context, src Source, peer tg.InputPeerClass) (*CachedPayload, error) {
	switch src.Type {
	case "channel", "supergroup":
		ch, ok := peer.(*tg.InputPeerChannel)
		if !ok {
			return nil, fmt.Errorf("channel source needs InputPeerChannel, got %T", peer)
		}
		res, err := retry.Do(ctx, s.retry, func() (*tg.MessagesChatFull, error) {
			r, err := s.api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
				ChannelID:  ch.ChannelID,
				AccessHash: ch.AccessHash,
			})
			return r, retry.ClassifyError(err)
		})
		if err != nil {
			return nil, fmt.Errorf("channels.getFullChannel: %w", err)
		}
		full, ok := res.FullChat.(*tg.ChannelFull)
		if !ok {
			return nil, fmt.Errorf("unexpected full chat type: %T", res.FullChat)
		}
		p := &CachedPayload{
			MembersCount: full.ParticipantsCount,
			Description:  full.About,
		}
		if linked, ok := full.GetLinkedChatID(); ok && linked != 0 {
			p.LinkedChatID = botAPIChannelID(linked)
			p.HasComments = true
		}
		if inv, ok := full.GetExportedInvite(); ok {
			if ex, ok := inv.(*tg.ChatInviteExported); ok {
				p.InviteLink = ex.Link
			}
		}
		// Creation date lives on the matching tg.Channel inside Chats, not
		// on ChannelFull itself.
		for _, c := range res.Chats {
			ch, ok := c.(*tg.Channel)
			if !ok || ch.ID != full.ID {
				continue
			}
			if ch.Date != 0 {
				p.CreationDate = time.Unix(int64(ch.Date), 0).UTC().Format(time.RFC3339)
			}
			break
		}
		return p, nil

	case "group":
		chatPeer, ok := peer.(*tg.InputPeerChat)
		if !ok {
			return nil, fmt.Errorf("group source needs InputPeerChat, got %T", peer)
		}
		res, err := retry.Do(ctx, s.retry, func() (*tg.MessagesChatFull, error) {
			r, err := s.api.MessagesGetFullChat(ctx, chatPeer.ChatID)
			return r, retry.ClassifyError(err)
		})
		if err != nil {
			return nil, fmt.Errorf("messages.getFullChat: %w", err)
		}
		full, ok := res.FullChat.(*tg.ChatFull)
		if !ok {
			return nil, fmt.Errorf("unexpected full chat type: %T", res.FullChat)
		}
		p := &CachedPayload{Description: full.About}
		if inv, ok := full.GetExportedInvite(); ok {
			if ex, ok := inv.(*tg.ChatInviteExported); ok {
				p.InviteLink = ex.Link
			}
		}
		for _, c := range res.Chats {
			ch, ok := c.(*tg.Chat)
			if !ok || ch.ID != full.ID {
				continue
			}
			if ch.Date != 0 {
				p.CreationDate = time.Unix(int64(ch.Date), 0).UTC().Format(time.RFC3339)
			}
			break
		}
		return p, nil

	case "user", "bot":
		var inputUser tg.InputUserClass
		switch p := peer.(type) {
		case *tg.InputPeerUser:
			inputUser = &tg.InputUser{UserID: p.UserID, AccessHash: p.AccessHash}
		case *tg.InputPeerSelf:
			inputUser = &tg.InputUserSelf{}
		default:
			return nil, fmt.Errorf("user source needs InputPeerUser or InputPeerSelf, got %T", peer)
		}
		res, err := retry.Do(ctx, s.retry, func() (*tg.UsersUserFull, error) {
			r, err := s.api.UsersGetFullUser(ctx, inputUser)
			return r, retry.ClassifyError(err)
		})
		if err != nil {
			return nil, fmt.Errorf("users.getFullUser: %w", err)
		}
		p := &CachedPayload{}
		if about, ok := res.FullUser.GetAbout(); ok {
			p.Description = about
		}
		return p, nil
	}
	return &CachedPayload{}, nil
}
