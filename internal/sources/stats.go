package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"

	"github.com/searchtgcli/tgs/internal/retry"
)

// fetchStats performs three Telegram calls to compute total/24h/first metrics
// for a single peer. Calls are sequential to minimise FLOOD_WAIT risk.
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

	since := int(time.Now().Add(-24 * time.Hour).Unix())
	dayRes, err := retry.Do(ctx, s.retry, func() (tg.MessagesMessagesClass, error) {
		r, err := s.api.MessagesSearch(ctx, &tg.MessagesSearchRequest{
			Peer:    peer,
			Q:       "",
			Filter:  &tg.InputMessagesFilterEmpty{},
			MinDate: since,
			Limit:   1,
		})
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.search (24h): %w", err)
	}
	day := extractCount(dayRes)

	stats := &Stats{TotalMessages: total, Messages24h: day}

	histRes, err := retry.Do(ctx, s.retry, func() (tg.MessagesMessagesClass, error) {
		r, err := s.api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:      peer,
			OffsetID:  1,
			AddOffset: -1,
			Limit:     1,
		})
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.getHistory (first): %w", err)
	}
	if m := firstMessage(histRes); m != nil {
		stats.FirstMessage = &FirstMessage{
			ID:   m.ID,
			Date: time.Unix(int64(m.Date), 0).UTC().Format(time.RFC3339),
		}
	}
	return stats, nil
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
