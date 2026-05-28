package sources

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)

func TestFetchStats_AllThreeQueries(t *testing.T) {
	now := int(time.Now().Unix())
	var calls []string
	api := &mockAPI{
		search: func(_ context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			calls = append(calls, "total")
			return &tg.MessagesMessagesSlice{Count: 9999}, nil
		},
		getHistory: func(_ context.Context, req *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			// Distinguish 24h paging (OffsetID=0, AddOffset=0) from first-msg
			// lookup (OffsetID=0, AddOffset=-1).
			if req.AddOffset == -1 {
				calls = append(calls, "first")
				return &tg.MessagesMessagesSlice{
					Messages: []tg.MessageClass{&tg.Message{ID: 1, Date: 1500000000}},
				}, nil
			}
			calls = append(calls, "24h")
			// 4 fresh + 1 stale → loop exits, count=4.
			return &tg.MessagesMessagesSlice{Messages: []tg.MessageClass{
				&tg.Message{ID: 10, Date: now - 60},
				&tg.Message{ID: 9, Date: now - 120},
				&tg.Message{ID: 8, Date: now - 180},
				&tg.Message{ID: 7, Date: now - 240},
				&tg.Message{ID: 6, Date: now - 48*60*60}, // stale → stop
			}}, nil
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
		t.Errorf("Messages24h = %d, want 4", got.Messages24h)
	}
	if got.FirstMessage == nil || got.FirstMessage.ID != 1 {
		t.Errorf("FirstMessage = %+v", got.FirstMessage)
	}
	if got.FirstMessage.Date != "2017-07-14T02:40:00Z" {
		t.Errorf("FirstMessage.Date = %q", got.FirstMessage.Date)
	}
	if len(calls) != 3 {
		t.Errorf("expected 3 calls (total, 24h, first), got %v", calls)
	}
}

// TestFetchStats_Messages24hCountsViaHistory verifies that the 24-hour count
// is computed by walking messages.getHistory (where MinDate is actually
// honoured) rather than messages.search with empty Q (where Telegram returns
// the total count, ignoring MinDate — the bug from P0-3).
func TestFetchStats_Messages24hCountsViaHistory(t *testing.T) {
	now := int(time.Now().Unix())
	day := 24 * 60 * 60
	api := &mockAPI{
		search: func(_ context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			if req.MinDate != 0 {
				t.Errorf("messages.search should not be used for 24h count anymore (MinDate=%d)", req.MinDate)
			}
			return &tg.MessagesMessagesSlice{Count: 1000}, nil // total only
		},
		getHistory: func(_ context.Context, req *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			// First call: latest page (OffsetID=0). Return 5 messages, 3
			// within the last 24h and 2 older.
			if req.OffsetID == 0 && req.AddOffset == 0 {
				return &tg.MessagesChannelMessages{
					Count: 1000,
					Messages: []tg.MessageClass{
						&tg.Message{ID: 100, Date: now - 60},   // fresh
						&tg.Message{ID: 99, Date: now - 3600},  // fresh
						&tg.Message{ID: 98, Date: now - 7200},  // fresh
						&tg.Message{ID: 97, Date: now - 2*day}, // stale
						&tg.Message{ID: 96, Date: now - 3*day}, // stale
					},
				}, nil
			}
			// First-message lookup (separate code path).
			return &tg.MessagesChannelMessages{
				Messages: []tg.MessageClass{&tg.Message{ID: 1, Date: now - 100*day}},
			}, nil
		},
	}
	s := &Service{api: api}
	got, err := s.fetchStats(context.Background(), &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1})
	if err != nil {
		t.Fatalf("fetchStats: %v", err)
	}
	if got.TotalMessages != 1000 {
		t.Errorf("TotalMessages = %d, want 1000", got.TotalMessages)
	}
	if got.Messages24h != 3 {
		t.Errorf("Messages24h = %d, want 3 (only those with date > now-24h)", got.Messages24h)
	}
}

// TestFetchStats_Messages24hPaginatesUntilStale verifies that when the first
// page is entirely within the 24h window, fetchStats keeps walking history
// until it crosses the cutoff.
func TestFetchStats_Messages24hPaginatesUntilStale(t *testing.T) {
	now := int(time.Now().Unix())
	day := 24 * 60 * 60
	call := 0
	api := &mockAPI{
		search: func(_ context.Context, _ *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error) {
			return &tg.MessagesMessagesSlice{Count: 200}, nil
		},
		getHistory: func(_ context.Context, req *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error) {
			call++
			switch call {
			case 1:
				// Page 1: 3 fresh messages, no stale yet.
				return &tg.MessagesChannelMessages{
					Messages: []tg.MessageClass{
						&tg.Message{ID: 30, Date: now - 60},
						&tg.Message{ID: 29, Date: now - 120},
						&tg.Message{ID: 28, Date: now - 180},
					},
				}, nil
			case 2:
				// Page 2: paginated from OffsetID=28; 2 more fresh, 1 stale.
				if req.OffsetID != 28 {
					t.Errorf("expected OffsetID=28 on page 2, got %d", req.OffsetID)
				}
				return &tg.MessagesChannelMessages{
					Messages: []tg.MessageClass{
						&tg.Message{ID: 27, Date: now - 86000},
						&tg.Message{ID: 26, Date: now - 86300},
						&tg.Message{ID: 25, Date: now - 2*day}, // stale → stop
					},
				}, nil
			default:
				// first_message lookup is allowed.
				return &tg.MessagesChannelMessages{
					Messages: []tg.MessageClass{&tg.Message{ID: 1, Date: now - 100*day}},
				}, nil
			}
		},
	}
	s := &Service{api: api}
	got, err := s.fetchStats(context.Background(), &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1})
	if err != nil {
		t.Fatalf("fetchStats: %v", err)
	}
	if got.Messages24h != 5 {
		t.Errorf("Messages24h = %d, want 5 (3 from page 1 + 2 from page 2 before stale)", got.Messages24h)
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
