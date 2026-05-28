package sources

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/searchtgcli/tgs/internal/resolve"
	"github.com/searchtgcli/tgs/internal/retry"
)

// Service exposes high-level operations over Telegram source data.
type Service struct {
	api      API
	resolver *resolve.Resolver
	retry    *retry.Policy
	cache    *SourceStatsCache
	selfID   int64
}

// New constructs a Service. resolver, retryPolicy and cache may be nil.
func New(api API, resolver *resolve.Resolver, retryPolicy *retry.Policy, cache *SourceStatsCache, selfID int64) *Service {
	if retryPolicy == nil {
		retryPolicy = retry.DefaultPolicy()
	}
	return &Service{
		api:      api,
		resolver: resolver,
		retry:    retryPolicy,
		cache:    cache,
		selfID:   selfID,
	}
}

// List returns sources from the user's dialogs, optionally enriched with
// expensive metrics (--with-stats).
func (s *Service) List(ctx context.Context, req ListRequest) (*ListResult, error) {
	cur, err := DecodeCursor(req.Cursor)
	if err != nil {
		return nil, err
	}

	pageSize := 100
	if req.Limit > 0 && req.Limit < pageSize {
		pageSize = req.Limit
	}

	var all []sourceWithPeer
	var totalAcc int
	var nextCursor *Cursor
	for {
		page, next, total, err := s.fetchDialogs(ctx, cur, pageSize, req.Archived, s.selfID)
		if err != nil {
			return nil, err
		}
		if total > totalAcc {
			totalAcc = total
		}
		all = append(all, page...)
		if next == nil {
			break
		}
		cur = next
		if req.Limit > 0 && len(all) >= req.Limit {
			nextCursor = next
			break
		}
	}
	if req.Limit > 0 && len(all) > req.Limit {
		all = all[:req.Limit]
	}

	filtered := filterByType(all, req.Types)

	if req.WithStats {
		s.enrichWithStats(ctx, filtered)
	}

	sources := make([]Source, len(filtered))
	for i, it := range filtered {
		sources[i] = it.Source
	}
	result := &ListResult{
		Sources: sources,
		Total:   totalAcc,
	}
	if nextCursor != nil {
		result.Cursor = nextCursor.Encode()
	}
	return result, nil
}

// filterByType returns only items whose Source.Type is in keep. Empty keep -> all.
func filterByType(items []sourceWithPeer, keep []string) []sourceWithPeer {
	if len(keep) == 0 {
		return items
	}
	set := make(map[string]bool, len(keep))
	for _, k := range keep {
		set[k] = true
	}
	out := make([]sourceWithPeer, 0, len(items))
	for _, it := range items {
		if set[it.Source.Type] {
			out = append(out, it)
		}
	}
	return out
}

// Inspect is implemented in Task 10.
func (s *Service) Inspect(ctx context.Context, req InspectRequest) (*Source, error) {
	return nil, errors.New("not implemented yet")
}

const activeWindow = 7 * 24 * time.Hour

// enrichWithStats populates Stats and full-info fields on each item. Per-
// source errors are captured into Source.StatsError and never abort the whole
// call.
func (s *Service) enrichWithStats(ctx context.Context, items []sourceWithPeer) {
	for i := range items {
		if ctx.Err() != nil {
			return
		}
		src := &items[i].Source
		peer := items[i].Peer

		lastID := 0
		if src.LastMessage != nil {
			lastID = src.LastMessage.ID
		}

		isActive := isActiveSource(*src)
		if !isActive && s.cache != nil {
			if cached, hit, err := s.cache.Get(src.ID, lastID); err == nil && hit {
				applyPayload(src, cached)
				continue
			}
		}

		payload, err := s.fetchFull(ctx, *src, peer)
		if err != nil {
			src.StatsError = telegramErrorCode(err)
			continue
		}
		stats, err := s.fetchStats(ctx, peer)
		if err != nil {
			src.StatsError = telegramErrorCode(err)
			applyPayload(src, payload)
			continue
		}
		payload.Stats = stats
		applyPayload(src, payload)

		if s.cache != nil && !isActive {
			_ = s.cache.Put(src.ID, lastID, payload)
		}
	}
}

func isActiveSource(src Source) bool {
	if src.LastMessage == nil {
		return false
	}
	d, err := time.Parse(time.RFC3339, src.LastMessage.Date)
	if err != nil {
		return true // fall back to live if we can't parse
	}
	return time.Since(d) < activeWindow
}

// applyPayload copies cached or freshly-fetched expensive fields onto a Source.
func applyPayload(src *Source, p *CachedPayload) {
	if p == nil {
		return
	}
	src.MembersCount = p.MembersCount
	if p.Description != "" {
		src.Description = p.Description
	}
	if p.LinkedChatID != 0 {
		src.LinkedChatID = p.LinkedChatID
		src.HasComments = p.HasComments
	}
	if p.InviteLink != "" {
		src.InviteLink = p.InviteLink
	}
	if p.CreationDate != "" {
		src.CreationDate = p.CreationDate
	}
	if p.Stats != nil {
		src.Stats = p.Stats
	}
}

// telegramErrorCode extracts a short uppercase code (e.g. CHANNEL_PRIVATE)
// from a gotd error message; falls back to "ERROR".
func telegramErrorCode(err error) string {
	msg := err.Error()
	// gotd errors usually look like: "...: rpc error code 400: CHANNEL_PRIVATE (...)".
	if idx := strings.LastIndex(msg, ": "); idx >= 0 {
		tail := msg[idx+2:]
		if sp := strings.Index(tail, " "); sp > 0 {
			tail = tail[:sp]
		}
		if tail != "" && strings.ToUpper(tail) == tail {
			return tail
		}
	}
	return "ERROR"
}
