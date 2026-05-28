package sources

import (
	"context"
	"errors"

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

	// --with-stats enrichment is added in Task 9.

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
