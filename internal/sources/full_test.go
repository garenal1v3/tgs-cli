package sources

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

func TestFetchFull_Channel(t *testing.T) {
	api := &mockAPI{
		getFullChannel: func(_ context.Context, _ tg.InputChannelClass) (*tg.MessagesChatFull, error) {
			full := &tg.ChannelFull{
				ID:                111,
				About:             "official channel",
				ParticipantsCount: 12345,
			}
			full.SetLinkedChatID(999)
			full.SetExportedInvite(&tg.ChatInviteExported{Link: "https://t.me/+abc"})
			return &tg.MessagesChatFull{FullChat: full}, nil
		},
	}
	s := &Service{api: api}
	got, err := s.fetchFull(context.Background(), Source{ID: -1000000000111, Type: "channel"}, &tg.InputPeerChannel{ChannelID: 111, AccessHash: 1})
	if err != nil {
		t.Fatalf("fetchFull: %v", err)
	}
	if got.MembersCount != 12345 {
		t.Errorf("MembersCount = %d", got.MembersCount)
	}
	if got.Description != "official channel" {
		t.Errorf("Description = %q", got.Description)
	}
	if got.LinkedChatID != -1000000000999 {
		t.Errorf("LinkedChatID = %d", got.LinkedChatID)
	}
	if !got.HasComments {
		t.Error("HasComments should be true when LinkedChatID is set")
	}
	if got.InviteLink != "https://t.me/+abc" {
		t.Errorf("InviteLink = %q", got.InviteLink)
	}
}

func TestFetchFull_LegacyGroup(t *testing.T) {
	api := &mockAPI{
		getFullChat: func(_ context.Context, _ int64) (*tg.MessagesChatFull, error) {
			full := &tg.ChatFull{ID: 5, About: "old"}
			return &tg.MessagesChatFull{FullChat: full}, nil
		},
	}
	s := &Service{api: api}
	got, err := s.fetchFull(context.Background(), Source{ID: -5, Type: "group"}, &tg.InputPeerChat{ChatID: 5})
	if err != nil {
		t.Fatalf("fetchFull: %v", err)
	}
	if got.Description != "old" {
		t.Errorf("Description = %q", got.Description)
	}
}

func TestFetchFull_User(t *testing.T) {
	api := &mockAPI{
		getFullUser: func(_ context.Context, _ tg.InputUserClass) (*tg.UsersUserFull, error) {
			full := &tg.UserFull{ID: 10}
			full.SetAbout("bio here")
			return &tg.UsersUserFull{FullUser: *full}, nil
		},
	}
	s := &Service{api: api}
	got, err := s.fetchFull(context.Background(), Source{ID: 10, Type: "user"}, &tg.InputPeerUser{UserID: 10, AccessHash: 1})
	if err != nil {
		t.Fatalf("fetchFull: %v", err)
	}
	if got.Description != "bio here" {
		t.Errorf("Description = %q", got.Description)
	}
}
