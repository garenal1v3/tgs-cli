package search

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotd/td/tg"

	"github.com/garenal1v3/tgs-cli/internal/retry"
)

// CalendarRequest describes a request to get search results grouped by date.
type CalendarRequest struct {
	Peer   tg.InputPeerClass
	Filter tg.MessagesFilterClass
}

// GetCalendar returns search results grouped by date for the given peer and filter.
func (s *Service) GetCalendar(ctx context.Context, req CalendarRequest) (*CalendarResult, error) {
	if req.Peer == nil {
		return nil, errors.New("peer is required")
	}
	if req.Filter == nil {
		return nil, errors.New("filter is required")
	}

	apiReq := &tg.MessagesGetSearchResultsCalendarRequest{
		Peer:   req.Peer,
		Filter: req.Filter,
	}

	res, err := retry.Do(ctx, s.retry, func() (*tg.MessagesSearchResultsCalendar, error) {
		r, err := s.api.MessagesGetSearchResultsCalendar(ctx, apiReq)
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.getSearchResultsCalendar: %w", err)
	}

	periods := make([]CalendarPeriod, 0, len(res.Periods))
	for _, p := range res.Periods {
		periods = append(periods, CalendarPeriod{
			Date:     time.Unix(int64(p.Date), 0).UTC().Format("2006-01-02"),
			Count:    p.Count,
			MinMsgID: p.MinMsgID,
			MaxMsgID: p.MaxMsgID,
		})
	}

	return &CalendarResult{
		Periods: periods,
		Total:   res.Count,
	}, nil
}

// PerChatCalendar pairs a chat reference with its calendar periods.
type PerChatCalendar struct {
	Chat    ChatRef          `json:"chat"`
	Periods []CalendarPeriod `json:"periods"`
	Total   int              `json:"total"`
}

// DateCount is one aggregated date bucket across all chats in a fan-out.
type DateCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// MultiCalendarResult is the response shape for multi-chat (folder) queries.
// Chats holds the per-chat breakdown; Totals aggregates the message count per
// date across every chat (newest date first); Total is the grand total across
// all chats and dates.
type MultiCalendarResult struct {
	Chats  []PerChatCalendar `json:"chats"`
	Totals []DateCount       `json:"totals"`
	Total  int               `json:"total"`
}
