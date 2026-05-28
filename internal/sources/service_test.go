package sources

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

func TestService_List_BasicFiltering(t *testing.T) {
	api := &mockAPI{
		getDialogs: func(_ context.Context, _ *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error) {
			channel := &tg.Channel{ID: 1, Title: "Ch", Broadcast: true}
			channel.SetAccessHash(1)
			group := &tg.Chat{ID: 2, Title: "Grp"}
			user := &tg.User{ID: 3}
			user.SetFirstName("Alice")
			user.SetAccessHash(3)
			return &tg.MessagesDialogs{
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 1}, TopMessage: 10},
					&tg.Dialog{Peer: &tg.PeerChat{ChatID: 2}, TopMessage: 20},
					&tg.Dialog{Peer: &tg.PeerUser{UserID: 3}, TopMessage: 30},
				},
				Messages: []tg.MessageClass{
					&tg.Message{ID: 10, Date: 1700000000},
					&tg.Message{ID: 20, Date: 1700000000},
					&tg.Message{ID: 30, Date: 1700000000},
				},
				Chats: []tg.ChatClass{channel, group},
				Users: []tg.UserClass{user},
			}, nil
		},
	}
	s := New(api, nil, nil, nil, 0)

	got, err := s.List(context.Background(), ListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Sources) != 3 {
		t.Errorf("len = %d, want 3", len(got.Sources))
	}

	// Filter
	got, err = s.List(context.Background(), ListRequest{Types: []string{"channel"}})
	if err != nil {
		t.Fatalf("List filter: %v", err)
	}
	if len(got.Sources) != 1 || got.Sources[0].Type != "channel" {
		t.Errorf("filtered = %+v", got.Sources)
	}
}
