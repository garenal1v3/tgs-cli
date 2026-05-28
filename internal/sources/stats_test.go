package sources

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

func TestFetchStats_AllThreeQueries(t *testing.T) {
	var calls []string
	api := &mockAPI{
		search: func(_ context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			if req.MinDate == 0 {
				calls = append(calls, "total")
				return &tg.MessagesMessagesSlice{Count: 9999}, nil
			}
			calls = append(calls, "24h")
			return &tg.MessagesMessagesSlice{Count: 4}, nil
		},
		getHistory: func(_ context.Context, _ *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			calls = append(calls, "first")
			return &tg.MessagesMessagesSlice{
				Messages: []tg.MessageClass{
					&tg.Message{ID: 1, Date: 1500000000},
				},
				Count: 9999,
			}, nil
		},
	}
	s := &Service{api: api}
	got, err := s.fetchStats(context.Background(), &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1})
	if err != nil {
		t.Fatalf("fetchStats: %v", err)
	}
	if got.TotalMessages != 9999 {
		t.Errorf("TotalMessages = %d", got.TotalMessages)
	}
	if got.Messages24h != 4 {
		t.Errorf("Messages24h = %d", got.Messages24h)
	}
	if got.FirstMessage == nil || got.FirstMessage.ID != 1 {
		t.Errorf("FirstMessage = %+v", got.FirstMessage)
	}
	if got.FirstMessage.Date != "2017-07-14T02:40:00Z" {
		t.Errorf("FirstMessage.Date = %q", got.FirstMessage.Date)
	}
	if len(calls) != 3 {
		t.Errorf("expected 3 calls, got %v", calls)
	}
}

func TestFetchStats_NoFirstMessage(t *testing.T) {
	api := &mockAPI{
		search: func(_ context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			return &tg.MessagesMessagesSlice{Count: 0}, nil
		},
		getHistory: func(_ context.Context, _ *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			return &tg.MessagesMessagesSlice{Messages: nil}, nil
		},
	}
	s := &Service{api: api}
	got, err := s.fetchStats(context.Background(), &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1})
	if err != nil {
		t.Fatalf("fetchStats: %v", err)
	}
	if got.FirstMessage != nil {
		t.Errorf("FirstMessage should be nil when no history, got %+v", got.FirstMessage)
	}
}
