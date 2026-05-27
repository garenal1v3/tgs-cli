package search

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotd/td/tg"

	"github.com/searchtgcli/tgs/internal/retry"
)

// CountersRequest describes a request to get message counts by filter type.
type CountersRequest struct {
	Peer    tg.InputPeerClass
	TopicID int
	Filters []tg.MessagesFilterClass
}

// GetCounters returns message counts for each filter type in the given peer.
func (s *Service) GetCounters(ctx context.Context, req CountersRequest) (*CountersResult, error) {
	if req.Peer == nil {
		return nil, errors.New("peer is required")
	}
	if len(req.Filters) == 0 {
		req.Filters = allFilters()
	}

	apiReq := &tg.MessagesGetSearchCountersRequest{
		Peer:    req.Peer,
		Filters: req.Filters,
	}
	if req.TopicID != 0 {
		apiReq.SetTopMsgID(req.TopicID)
	}

	counters, err := retry.Do(ctx, s.retry, func() ([]tg.MessagesSearchCounter, error) {
		r, err := s.api.MessagesGetSearchCounters(ctx, apiReq)
		return r, retry.ClassifyError(err)
	})
	if err != nil {
		return nil, fmt.Errorf("messages.getSearchCounters: %w", err)
	}

	entries := make([]CounterEntry, 0, len(counters))
	for _, c := range counters {
		name := filterToName(c.Filter)
		if name == "" {
			continue
		}
		entries = append(entries, CounterEntry{
			Filter: name,
			Count:  c.Count,
		})
	}

	return &CountersResult{Counters: entries}, nil
}

// allFilters returns a slice of all 14 non-empty filter instances
// (one per filter type supported by Telegram).
func allFilters() []tg.MessagesFilterClass {
	return []tg.MessagesFilterClass{
		&tg.InputMessagesFilterPhotos{},
		&tg.InputMessagesFilterVideo{},
		&tg.InputMessagesFilterPhotoVideo{},
		&tg.InputMessagesFilterDocument{},
		&tg.InputMessagesFilterURL{},
		&tg.InputMessagesFilterGif{},
		&tg.InputMessagesFilterVoice{},
		&tg.InputMessagesFilterMusic{},
		&tg.InputMessagesFilterRoundVideo{},
		&tg.InputMessagesFilterGeo{},
		&tg.InputMessagesFilterContacts{},
		&tg.InputMessagesFilterPinned{},
		&tg.InputMessagesFilterMyMentions{},
		&tg.InputMessagesFilterPhoneCalls{},
	}
}

// filterToName maps a MessagesFilterClass back to its string name.
func filterToName(f tg.MessagesFilterClass) string {
	switch f.(type) {
	case *tg.InputMessagesFilterPhotos:
		return "photo"
	case *tg.InputMessagesFilterVideo:
		return "video"
	case *tg.InputMessagesFilterPhotoVideo:
		return "photo-video"
	case *tg.InputMessagesFilterDocument:
		return "document"
	case *tg.InputMessagesFilterURL:
		return "url"
	case *tg.InputMessagesFilterGif:
		return "gif"
	case *tg.InputMessagesFilterVoice:
		return "voice"
	case *tg.InputMessagesFilterMusic:
		return "music"
	case *tg.InputMessagesFilterRoundVideo:
		return "round-video"
	case *tg.InputMessagesFilterGeo:
		return "geo"
	case *tg.InputMessagesFilterContacts:
		return "contact"
	case *tg.InputMessagesFilterPinned:
		return "pinned"
	case *tg.InputMessagesFilterMyMentions:
		return "mention"
	case *tg.InputMessagesFilterPhoneCalls:
		return "phone-call"
	case *tg.InputMessagesFilterChatPhotos:
		return "chat-photo"
	default:
		return ""
	}
}
