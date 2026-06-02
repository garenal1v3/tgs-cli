package search

import (
	"testing"

	"github.com/gotd/td/tg"

	"github.com/garenal1v3/tgs-cli/internal/sources"
)

func TestChatRefFromSource_CarriesDisplayFields(t *testing.T) {
	ref := chatRefFromSource(sources.Source{
		ID:       -1001234567890,
		Type:     "channel",
		Title:    "News",
		Username: "newschan",
	})
	if ref.ID != -1001234567890 || ref.Type != "channel" || ref.Title != "News" || ref.Username != "newschan" {
		t.Errorf("ref = %+v, want full display fields preserved", ref)
	}
}

func TestChatRefFromPeer_BotAPIIDs(t *testing.T) {
	tests := []struct {
		name string
		peer tg.InputPeerClass
		want int64
		typ  string
	}{
		{"channel", &tg.InputPeerChannel{ChannelID: 1234567890}, -1001234567890, ""},
		{"chat", &tg.InputPeerChat{ChatID: 42}, -42, "group"},
		{"user", &tg.InputPeerUser{UserID: 777}, 777, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := chatRefFromPeer(tt.peer)
			if ref.ID != tt.want {
				t.Errorf("ID = %d, want %d", ref.ID, tt.want)
			}
			if ref.Type != tt.typ {
				t.Errorf("Type = %q, want %q", ref.Type, tt.typ)
			}
		})
	}
}
