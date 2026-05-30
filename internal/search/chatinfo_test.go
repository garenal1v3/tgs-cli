package search

import (
	"testing"

	"github.com/gotd/td/tg"
)

// buildChatMap must emit Bot-API chat IDs (negative for channels/groups) while
// keying the map by the raw Telegram ID, so search output is consistent with
// `sources` and the IDs round-trip into --chat / --folder.
func TestBuildChatMap_BotAPIIDs(t *testing.T) {
	chats := []tg.ChatClass{
		&tg.Channel{ID: 1972880113, Title: "Chan", Broadcast: true},
		&tg.Channel{ID: 2002447117, Title: "Super"},
		&tg.Chat{ID: 42, Title: "LegacyGroup"},
	}
	m := buildChatMap(chats)

	if got := m[1972880113].ID; got != -1001972880113 {
		t.Errorf("channel ID = %d, want -1001972880113", got)
	}
	if got := m[1972880113].Type; got != "channel" {
		t.Errorf("channel Type = %q, want channel", got)
	}
	if got := m[2002447117].ID; got != -1002002447117 {
		t.Errorf("supergroup ID = %d, want -1002002447117", got)
	}
	if got := m[2002447117].Type; got != "supergroup" {
		t.Errorf("supergroup Type = %q, want supergroup", got)
	}
	if got := m[42].ID; got != -42 {
		t.Errorf("legacy group ID = %d, want -42", got)
	}
}

func TestBotAPIPeerID(t *testing.T) {
	tests := []struct {
		peer tg.PeerClass
		want int64
	}{
		{&tg.PeerUser{UserID: 777}, 777},
		{&tg.PeerChat{ChatID: 42}, -42},
		{&tg.PeerChannel{ChannelID: 1972880113}, -1001972880113},
	}
	for _, tt := range tests {
		if got := botAPIPeerID(tt.peer); got != tt.want {
			t.Errorf("botAPIPeerID(%T) = %d, want %d", tt.peer, got, tt.want)
		}
	}
}
