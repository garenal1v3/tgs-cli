package search

import (
	"context"

	"github.com/gotd/td/tg"
)

// resolveLinkedChats walks the given peers; for every channel peer with a
// linked discussion group, it returns an InputPeer for that group. Errors for
// individual channels are silently skipped (e.g. private channels we can't
// inspect) so a single failure does not abort the whole search.
func resolveLinkedChats(ctx context.Context, api *tg.Client, peers []tg.InputPeerClass) []tg.InputPeerClass {
	seen := make(map[int64]bool, len(peers))
	for _, p := range peers {
		if ch, ok := p.(*tg.InputPeerChannel); ok {
			seen[ch.ChannelID] = true
		}
	}

	var linked []tg.InputPeerClass
	for _, p := range peers {
		ch, ok := p.(*tg.InputPeerChannel)
		if !ok {
			continue
		}
		inputCh := &tg.InputChannel{
			ChannelID:  ch.ChannelID,
			AccessHash: ch.AccessHash,
		}
		full, err := api.ChannelsGetFullChannel(ctx, inputCh)
		if err != nil {
			continue
		}
		channelFull, ok := full.FullChat.(*tg.ChannelFull)
		if !ok {
			continue
		}
		linkedID, ok := channelFull.GetLinkedChatID()
		if !ok || linkedID == 0 || seen[linkedID] {
			continue
		}
		for _, c := range full.Chats {
			if chCh, ok := c.(*tg.Channel); ok && chCh.ID == linkedID {
				accessHash, _ := chCh.GetAccessHash()
				linked = append(linked, &tg.InputPeerChannel{
					ChannelID:  linkedID,
					AccessHash: accessHash,
				})
				seen[linkedID] = true
				break
			}
		}
	}
	return linked
}
