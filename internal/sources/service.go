package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/tg"

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

// Inspect resolves ref via the Resolver (using ResolveWithMeta to capture
// subscription state for channels) then fetches full info + (optionally)
// stats. Supports refs the user is not subscribed to (channels/users by
// @username or phone).
func (s *Service) Inspect(ctx context.Context, req InspectRequest) (*Source, error) {
	if s.resolver == nil {
		return nil, errors.New("inspect requires a resolver")
	}
	ref := strings.TrimSpace(req.Ref)

	peer, subscribed, src, err := s.resolveForInspect(ctx, ref)
	if err != nil {
		return nil, err
	}
	return s.inspectKnownWithSubscription(ctx, src, peer, req.NoStats, subscribed)
}

// resolveForInspect resolves ref to an InputPeer, an initial Source seed (with
// type/title/etc), and a *bool indicating subscription state. Returns
// subscribed=nil for users/bots (subscription concept doesn't apply) and for
// channels resolved from cache or by numeric ID (Left flag unknown — falls
// back to error-code detection in inspectKnownWithSubscription).
func (s *Service) resolveForInspect(ctx context.Context, ref string) (tg.InputPeerClass, *bool, Source, error) {
	// Self alias.
	if ref == "@me" || ref == "-" {
		if s.selfID == 0 {
			return nil, nil, Source{}, errors.New("self ID is unknown")
		}
		yes := true
		return &tg.InputPeerSelf{}, &yes, Source{
			ID: s.selfID, Type: "user", Saved: true, Title: "Saved Messages",
		}, nil
	}

	peer, meta, err := s.resolver.ResolveWithMeta(ctx, ref)
	if err != nil {
		return nil, nil, Source{}, err
	}

	// Build a seed Source from peer type. Concrete fields are filled by fetchFull.
	var src Source
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		src = Source{ID: botAPIChannelID(p.ChannelID), Type: "channel"}
	case *tg.InputPeerChat:
		src = Source{ID: -p.ChatID, Type: "group"}
	case *tg.InputPeerUser:
		src = Source{ID: p.UserID, Type: "user"}
	default:
		return nil, nil, Source{}, fmt.Errorf("unsupported peer type: %T", peer)
	}
	return peer, meta.Subscribed, src, nil
}

// inspectKnown is the entry point used by tests that supply a known peer.
// It sets subscribed=*true and delegates to inspectKnownWithSubscription.
func (s *Service) inspectKnown(ctx context.Context, src Source, peer tg.InputPeerClass, noStats bool) (*Source, error) {
	yes := true
	return s.inspectKnownWithSubscription(ctx, src, peer, noStats, &yes)
}

// inspectKnownWithSubscription completes a Source by pulling full info and
// (optionally) stats. If subscribed is *false, stats are skipped and set to nil.
//
// subscribed may be nil ("unknown") — in that case we still try fetchFull and
// only conclude "not subscribed" if the API returns an access-denied error.
func (s *Service) inspectKnownWithSubscription(ctx context.Context, src Source, peer tg.InputPeerClass, noStats bool, subscribed *bool) (*Source, error) {
	src.Subscribed = subscribed

	// Skip the network call entirely if we already know we're not subscribed.
	if subscribed != nil && !*subscribed {
		return &src, nil
	}

	payload, err := s.fetchFull(ctx, src, peer)
	if err == nil {
		applyPayload(&src, payload)
	} else {
		code := telegramErrorCode(err)
		src.StatsError = code
		// Fallback for subscribed==nil paths (cache hit, numeric ID): if
		// full-info fetch fails with a "no access" code, conclude not subscribed.
		if code == "CHANNEL_PRIVATE" || code == "USER_PRIVACY_RESTRICTED" || code == "CHANNEL_INVALID" {
			no := false
			src.Subscribed = &no
			return &src, nil
		}
	}

	if noStats {
		return &src, nil
	}

	stats, err := s.fetchStats(ctx, peer)
	if err != nil {
		// Don't overwrite an existing stats_error from full info.
		if src.StatsError == "" {
			src.StatsError = telegramErrorCode(err)
		}
		return &src, nil
	}
	src.Stats = stats
	return &src, nil
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
