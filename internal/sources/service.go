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
//
// When req.Archived is true, the list contains both main (folder 0) and
// archived (folder 1) dialogs — Telegram's messages.getDialogs returns one
// folder at a time, so we walk folder 0 first and then continue with folder
// 1. The cursor carries the current folder so paginated calls resume in the
// right place.
func (s *Service) List(ctx context.Context, req ListRequest) (*ListResult, error) {
	cur, err := DecodeCursor(req.Cursor)
	if err != nil {
		return nil, err
	}

	pageSize := 100
	if req.Limit > 0 && req.Limit < pageSize {
		pageSize = req.Limit
	}

	// Start folder is determined by the cursor (resuming pagination) or by 0
	// for a fresh call.
	currentFolder := 0
	if cur != nil {
		currentFolder = cur.Folder
	}

	var all []sourceWithPeer
	var totalAcc int
	var totalSeen int // unfiltered count actually returned across pages
	var nextCursor *Cursor
	for {
		page, next, total, err := s.fetchDialogs(ctx, cur, pageSize, currentFolder, s.selfID)
		if err != nil {
			return nil, err
		}
		if total > totalAcc {
			totalAcc = total
		}
		totalSeen += len(page)
		// Filter each page before accumulating so the limit check counts only
		// matching items (fixes: --type bot --limit 2 returning empty results).
		page = filterByType(page, req.Types)

		// Apply --limit BEFORE appending the page in full, so we don't
		// silently drop matched items from a long final page. Without this
		// pre-check we'd accumulate len(all) > req.Limit, then truncate the
		// tail at the end of the function — and any next-page cursor would
		// point past the dropped tail. Building the cursor from the
		// last-kept item instead lets a follow-up call resume exactly where
		// this one stopped, regardless of whether more items are available
		// in the current folder or the archive.
		if req.Limit > 0 && len(all)+len(page) >= req.Limit {
			take := req.Limit - len(all)
			if take > 0 {
				all = append(all, page[:take]...)
				nextCursor = cursorFromItem(all[len(all)-1], currentFolder)
			} else if next != nil {
				// We already had >= limit items from a previous page; pass
				// through the Telegram-supplied cursor for the next call.
				nextCursor = next
			}
			break
		}
		all = append(all, page...)

		if next != nil {
			cur = next
			continue
		}

		// Current folder exhausted. If --archived was requested and we're
		// still on the main folder, jump to the archive folder.
		if req.Archived && currentFolder == 0 {
			currentFolder = 1
			cur = nil
			continue
		}

		break
	}
	// Telegram's reported Count is sometimes smaller than the slice it
	// actually returns (stale dialog-count cache). Make sure total never
	// under-reports what we returned.
	if totalSeen > totalAcc {
		totalAcc = totalSeen
	}

	if req.WithStats {
		s.enrichWithStats(ctx, all)
	}

	sources := make([]Source, len(all))
	for i, it := range all {
		sources[i] = it.Source
	}
	result := &ListResult{
		Sources:  sources,
		Total:    totalAcc,
		Returned: len(sources),
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
// type/title/etc), and a *bool indicating subscription state. The seed Source
// is populated from the resolver snapshot (live API response or cached) so
// that title/username/access/verified are present even when fetchFull is
// skipped (unsubscribed, private, or network failure).
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

	// Numeric-ID resolves only succeed if a previous @username/+phone resolve
	// populated the peer cache. Without an access hash, Telegram's
	// channels.getFullChannel / users.getFullUser silently return garbage —
	// so reject the call up front with a hint about what to use instead.
	if meta.Snapshot == nil && needsAccessHash(peer) {
		return nil, nil, Source{}, fmt.Errorf(
			"cannot inspect by numeric ID %q without an access hash: peer is not in the local cache. Run `tgs sources inspect @<username>` (or `+<phone>`) once to cache the peer, then numeric-ID inspect will work",
			ref,
		)
	}

	src := buildInspectSeed(peer, meta.Snapshot)
	if src.ID == 0 {
		return nil, nil, Source{}, fmt.Errorf("unsupported peer type: %T", peer)
	}
	src = markSelf(src, s.selfID)
	subscribed := meta.Subscribed
	if src.Saved && subscribed == nil {
		yes := true
		subscribed = &yes
	}
	return peer, subscribed, src, nil
}

// cursorFromItem builds a getDialogs offset pointing AT the supplied item;
// Telegram then returns dialogs strictly older than (peer, date, id) on the
// next request. Used by List when --limit interrupts iteration mid-page —
// the Telegram-supplied next cursor would point past the tail we discarded,
// so we synthesise one from the last item we actually kept.
func cursorFromItem(it sourceWithPeer, folder int) *Cursor {
	c := &Cursor{Folder: folder}
	if it.Source.LastMessage != nil {
		c.OffsetID = it.Source.LastMessage.ID
		if t, err := time.Parse(time.RFC3339, it.Source.LastMessage.Date); err == nil {
			c.OffsetDate = int(t.Unix())
		}
	}
	switch p := it.Peer.(type) {
	case *tg.InputPeerChannel:
		c.OffsetPeerType, c.OffsetPeerID, c.OffsetPeerAccessHash = "channel", p.ChannelID, p.AccessHash
	case *tg.InputPeerChat:
		c.OffsetPeerType, c.OffsetPeerID = "chat", p.ChatID
	case *tg.InputPeerUser:
		c.OffsetPeerType, c.OffsetPeerID, c.OffsetPeerAccessHash = "user", p.UserID, p.AccessHash
	}
	return c
}

// needsAccessHash reports whether a peer needs a non-zero access hash to be
// useful for follow-up Telegram calls. Channels and users do; legacy chats
// (InputPeerChat) don't, and InputPeerSelf is always valid.
func needsAccessHash(peer tg.InputPeerClass) bool {
	switch p := peer.(type) {
	case *tg.InputPeerUser:
		return p.AccessHash == 0
	case *tg.InputPeerChannel:
		return p.AccessHash == 0
	}
	return false
}

// markSelf flags a Source as Saved Messages when its ID matches the
// authenticated user's selfID — so `tgs sources inspect +<own-phone>` is
// consistent with `inspect @me` / `inspect -`.
func markSelf(src Source, selfID int64) Source {
	if selfID == 0 || src.ID != selfID {
		return src
	}
	src.Saved = true
	if src.Title == "" {
		src.Title = "Saved Messages"
	}
	return src
}

// buildInspectSeed constructs the initial Source for an inspect call. When a
// snapshot is available (live resolve or cache hit with snapshot fields), it
// copies title/username/access/members/flags into the seed so that those
// fields are present in the response even when fetchFull is unavailable.
//
// With a nil snapshot — e.g. ID-based resolves with cache miss — only the
// peer type and ID are populated; fetchFull (when subscribed) will fill in
// the rest.
func buildInspectSeed(peer tg.InputPeerClass, snap *resolve.PeerSnapshot) Source {
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		src := Source{ID: botAPIChannelID(p.ChannelID), Type: "channel"}
		if snap != nil {
			if snap.Broadcast {
				src.Type = "channel"
			} else if snap.Title != "" || snap.Username != "" || snap.HasTopics || snap.Gigagroup || snap.MembersCount > 0 {
				// Snapshot present but Broadcast=false → supergroup.
				src.Type = "supergroup"
			}
			src.Title = snap.Title
			src.Username = snap.Username
			src.Access = snap.Access
			src.MembersCount = snap.MembersCount
			src.Verified = snap.Verified
			src.Scam = snap.Scam
			src.Fake = snap.Fake
			src.Restricted = snap.Restricted
			src.HasTopics = snap.HasTopics
			src.Gigagroup = snap.Gigagroup
		}
		return src
	case *tg.InputPeerChat:
		src := Source{ID: -p.ChatID, Type: "group"}
		if snap != nil {
			src.Title = snap.Title
			src.MembersCount = snap.MembersCount
		}
		return src
	case *tg.InputPeerUser:
		src := Source{ID: p.UserID, Type: "user"}
		if snap != nil {
			if snap.IsBot {
				src.Type = "bot"
			}
			src.FirstName = snap.FirstName
			src.LastName = snap.LastName
			src.Username = snap.Username
			src.Phone = snap.Phone
			src.Verified = snap.Verified
			src.Scam = snap.Scam
			src.Fake = snap.Fake
			src.Restricted = snap.Restricted
			src.Deleted = snap.Deleted
		}
		return src
	}
	return Source{}
}

// inspectKnown is the entry point used by tests that supply a known peer.
// It sets subscribed=*true and delegates to inspectKnownWithSubscription.
func (s *Service) inspectKnown(ctx context.Context, src Source, peer tg.InputPeerClass, noStats bool) (*Source, error) {
	yes := true
	return s.inspectKnownWithSubscription(ctx, src, peer, noStats, &yes)
}

// inspectKnownWithSubscription completes a Source by pulling full info and
// (optionally) stats.
//
//   - subscribed == *true  → fetchFull + fetchStats
//   - subscribed == *false → fetchFull (still works for public channels),
//     stats skipped (would just yield zeros without membership)
//   - subscribed == nil    → fetchFull; on CHANNEL_PRIVATE-class errors infer
//     subscribed=*false and stop, otherwise proceed with stats.
func (s *Service) inspectKnownWithSubscription(ctx context.Context, src Source, peer tg.InputPeerClass, noStats bool, subscribed *bool) (*Source, error) {
	src.Subscribed = subscribed

	payload, fullErr := s.fetchFull(ctx, src, peer)
	if fullErr == nil {
		applyPayload(&src, payload)
	} else {
		code := telegramErrorCode(fullErr)
		src.StatsError = code
		// CHANNEL_PRIVATE / no-access errors are a signal: when subscription
		// state was unknown, infer *false and stop.
		if subscribed == nil && (code == "CHANNEL_PRIVATE" || code == "USER_PRIVACY_RESTRICTED" || code == "CHANNEL_INVALID") {
			no := false
			src.Subscribed = &no
			return &src, nil
		}
	}

	// Known-unsubscribed: don't attempt stats — without membership they'd be
	// either rejected or return zeros.
	if subscribed != nil && !*subscribed {
		return &src, nil
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
	// Stats succeeded: clear any earlier fetchFull error since the peer is
	// clearly accessible and we have valid data.
	src.StatsError = ""
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
