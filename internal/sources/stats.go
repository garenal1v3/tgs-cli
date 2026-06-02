package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"

	"github.com/garenal1v3/tgs-cli/internal/retry"
)

// messages24hPageLimit caps how many history pages we walk to compute the
// 24h count. 10 pages × 100 messages = 1000 max — beyond that the value is
// reported as messages24hPageLimit*100 (an under-count for hyper-active
// peers).
const messages24hPageLimit = 10

// messages24hPageSize is the Limit value we ask for from getHistory when
// counting 24h messages. 100 is the typical max per call.
const messages24hPageSize = 100

// fetchStats computes total / 24h / first-message metrics for one peer.
// Calls are sequential to minimise FLOOD_WAIT risk.
//
//   - total uses messages.search with empty Q + InputMessagesFilterEmpty —
//     Telegram returns the exact dialog message count there.
//   - 24h walks messages.getHistory backwards from the latest message,
//     counting messages whose Date is within the last 24h. We stop as soon as
//     we cross the cutoff or hit messages24hPageLimit pages.
//   - first message uses messages.getHistory with OffsetID=0,AddOffset=-1,
//     Limit=1 which returns the oldest stored message (the previous
//     OffsetID=1 form returns nothing for large channels — P0-4).
func (s *Service) fetchStats(ctx context.Context, peer tg.InputPeerClass) (*Stats, error) {
	totalRes, err := retry.Do(ctx, s.retry, func() (tg.MessagesMessagesClass, error) {
		r, err := s.api.MessagesSearch(ctx, &tg.MessagesSearchRequest{
			Peer:   peer,
			Q:      "",
			Filter: &tg.InputMessagesFilterEmpty{},
			Limit:  1,
		})
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.search (total): %w", err)
	}
	total := extractCount(totalRes)

	day, err := s.count24h(ctx, peer)
	if err != nil {
		return nil, fmt.Errorf("count 24h: %w", err)
	}

	stats := &Stats{TotalMessages: total, Messages24h: day}

	first, err := s.fetchFirstMessage(ctx, peer)
	if err != nil {
		return nil, fmt.Errorf("messages.getHistory (first): %w", err)
	}
	if first != nil {
		stats.FirstMessage = first
	}
	return stats, nil
}

// count24h paginates messages.getHistory from the latest message, counting
// those within the last 24h. Stops as soon as a message older than the cutoff
// appears (history is date-descending) or after messages24hPageLimit pages.
func (s *Service) count24h(ctx context.Context, peer tg.InputPeerClass) (int, error) {
	cutoff := int(time.Now().Add(-24 * time.Hour).Unix())
	offsetID := 0
	count := 0
	for page := 0; page < messages24hPageLimit; page++ {
		req := &tg.MessagesGetHistoryRequest{
			Peer:     peer,
			OffsetID: offsetID,
			Limit:    messages24hPageSize,
		}
		res, err := retry.Do(ctx, s.retry, func() (tg.MessagesMessagesClass, error) {
			r, err := s.api.MessagesGetHistory(ctx, req)
			return r, retry.ClassifyError(err)
		})
		if err != nil {
			return 0, err
		}
		msgs := extractMessages(res)
		if len(msgs) == 0 {
			return count, nil
		}
		var oldest int
		for _, m := range msgs {
			if m.Date >= cutoff {
				count++
			}
			if oldest == 0 || m.ID < oldest {
				oldest = m.ID
			}
		}
		// Stop once the last message of the page is already older than cutoff
		// — further pages are guaranteed older.
		if msgs[len(msgs)-1].Date < cutoff {
			return count, nil
		}
		offsetID = oldest
	}
	return count, nil
}

// fetchFirstMessage returns the oldest message in the peer's history.
//
// Strategy: ask for messages older than the unix epoch + 1s using OffsetDate
// (so add_offset can step *back* one slot to reach the very first message in
// chronological order). For channels and supergroups this is the only
// reliable path — the old OffsetID=1, AddOffset=-1 form returned an empty
// slice for large broadcast channels (P0-4).
//
// If that returns nothing (very old peers cleared from storage, or a peer
// the API refuses), fall back to OffsetID=0,AddOffset=-1 which works for
// most user-to-user dialogs.
//
// nil result means "unknown / unavailable" and is encoded as the omitted
// first_message field in the JSON response.
func (s *Service) fetchFirstMessage(ctx context.Context, peer tg.InputPeerClass) (*FirstMessage, error) {
	if m, err := s.firstMessageVia(ctx, peer, &tg.MessagesGetHistoryRequest{
		Peer:       peer,
		OffsetDate: 1, // 1970-01-01T00:00:01Z → "older than epoch+1s"
		AddOffset:  -1,
		Limit:      1,
	}); err != nil {
		return nil, err
	} else if m != nil {
		return m, nil
	}
	// Fallback for peers where the OffsetDate form returns nothing.
	return s.firstMessageVia(ctx, peer, &tg.MessagesGetHistoryRequest{
		Peer:      peer,
		OffsetID:  0,
		AddOffset: -1,
		Limit:     1,
	})
}

func (s *Service) firstMessageVia(ctx context.Context, _ tg.InputPeerClass, req *tg.MessagesGetHistoryRequest) (*FirstMessage, error) {
	res, err := retry.Do(ctx, s.retry, func() (tg.MessagesMessagesClass, error) {
		r, err := s.api.MessagesGetHistory(ctx, req)
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, err
	}
	m := firstMessage(res)
	if m == nil {
		return nil, nil
	}
	return &FirstMessage{
		ID:   m.ID,
		Date: time.Unix(int64(m.Date), 0).UTC().Format(time.RFC3339),
	}, nil
}

// extractMessages flattens any of the messages.* response classes into a
// concrete *tg.Message slice (skipping service messages / empty messages).
func extractMessages(res tg.MessagesMessagesClass) []*tg.Message {
	var raw []tg.MessageClass
	switch v := res.(type) {
	case *tg.MessagesMessages:
		raw = v.Messages
	case *tg.MessagesMessagesSlice:
		raw = v.Messages
	case *tg.MessagesChannelMessages:
		raw = v.Messages
	}
	out := make([]*tg.Message, 0, len(raw))
	for _, mc := range raw {
		if m, ok := mc.(*tg.Message); ok {
			out = append(out, m)
		}
	}
	return out
}

func extractCount(res tg.MessagesMessagesClass) int {
	switch v := res.(type) {
	case *tg.MessagesMessages:
		return len(v.Messages)
	case *tg.MessagesMessagesSlice:
		return v.Count
	case *tg.MessagesChannelMessages:
		return v.Count
	}
	return 0
}

func firstMessage(res tg.MessagesMessagesClass) *tg.Message {
	var msgs []tg.MessageClass
	switch v := res.(type) {
	case *tg.MessagesMessages:
		msgs = v.Messages
	case *tg.MessagesMessagesSlice:
		msgs = v.Messages
	case *tg.MessagesChannelMessages:
		msgs = v.Messages
	}
	for _, mc := range msgs {
		if m, ok := mc.(*tg.Message); ok {
			return m
		}
	}
	return nil
}
