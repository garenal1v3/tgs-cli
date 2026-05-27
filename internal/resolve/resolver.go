package resolve

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

// API is the subset of the Telegram client used by the Resolver.
type API interface {
	ContactsResolveUsername(ctx context.Context, req *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error)
	ContactsResolvePhone(ctx context.Context, req *tg.ContactsResolvePhoneRequest) (*tg.ContactsResolvedPeer, error)
}

// Resolver resolves user-supplied peer references (usernames, phone numbers, IDs)
// into InputPeerClass values suitable for Telegram API calls.
type Resolver struct {
	api   API
	cache *PeerCache // can be nil (no caching)
}

// NewResolver creates a Resolver. cache may be nil to disable caching.
func NewResolver(api API, cache *PeerCache) *Resolver {
	return &Resolver{api: api, cache: cache}
}

// Resolve parses a single input string and returns the corresponding InputPeerClass.
func (r *Resolver) Resolve(ctx context.Context, input string) (tg.InputPeerClass, error) {
	pi, err := ParseInput(input)
	if err != nil {
		return nil, fmt.Errorf("parse input %q: %w", input, err)
	}

	switch pi.Type {
	case InputUsername:
		return r.resolveUsername(ctx, pi.Value)
	case InputPhone:
		return r.resolvePhone(ctx, pi.Value)
	case InputID:
		return idToPeer(pi.ID), nil
	case InputInvite:
		return nil, fmt.Errorf("invite links are not supported for search")
	default:
		return nil, fmt.Errorf("unknown input type: %d", pi.Type)
	}
}

// ResolveMulti resolves multiple inputs sequentially, returning a slice of peers
// in the same order as the inputs.
func (r *Resolver) ResolveMulti(ctx context.Context, inputs []string) ([]tg.InputPeerClass, error) {
	peers := make([]tg.InputPeerClass, 0, len(inputs))
	for _, input := range inputs {
		peer, err := r.Resolve(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", input, err)
		}
		peers = append(peers, peer)
	}
	return peers, nil
}

// resolveUsername resolves a username via cache (if available) or the Telegram API.
func (r *Resolver) resolveUsername(ctx context.Context, username string) (tg.InputPeerClass, error) {
	// Check cache first.
	if r.cache != nil {
		entry, found, err := r.cache.Load(username)
		if err != nil {
			return nil, fmt.Errorf("cache load %q: %w", username, err)
		}
		if found {
			return entryToPeer(entry), nil
		}
	}

	res, err := r.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve username %q: %w", username, err)
	}

	peer, entry, err := resolvedToPeer(res)
	if err != nil {
		return nil, err
	}

	// Store in cache.
	if r.cache != nil {
		if storeErr := r.cache.Store(username, entry); storeErr != nil {
			return nil, fmt.Errorf("cache store %q: %w", username, storeErr)
		}
	}

	return peer, nil
}

// resolvePhone resolves a phone number via the Telegram API.
func (r *Resolver) resolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	res, err := r.api.ContactsResolvePhone(ctx, &tg.ContactsResolvePhoneRequest{
		Phone: phone,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve phone %q: %w", phone, err)
	}

	peer, _, err := resolvedToPeer(res)
	if err != nil {
		return nil, err
	}

	return peer, nil
}

// resolvedToPeer extracts an InputPeerClass and CacheEntry from a ContactsResolvedPeer.
func resolvedToPeer(res *tg.ContactsResolvedPeer) (tg.InputPeerClass, CacheEntry, error) {
	switch p := res.Peer.(type) {
	case *tg.PeerUser:
		for _, u := range res.Users {
			user, ok := u.(*tg.User)
			if !ok {
				continue
			}
			if user.GetID() == p.UserID {
				accessHash, _ := user.GetAccessHash()
				entry := CacheEntry{
					PeerType:   "user",
					ID:         user.GetID(),
					AccessHash: accessHash,
					ResolvedAt: nowUnix(),
				}
				return &tg.InputPeerUser{
					UserID:     user.GetID(),
					AccessHash: accessHash,
				}, entry, nil
			}
		}
		return nil, CacheEntry{}, fmt.Errorf("user %d not found in resolved response", p.UserID)

	case *tg.PeerChat:
		entry := CacheEntry{
			PeerType:   "chat",
			ID:         p.ChatID,
			ResolvedAt: nowUnix(),
		}
		return &tg.InputPeerChat{ChatID: p.ChatID}, entry, nil

	case *tg.PeerChannel:
		for _, c := range res.Chats {
			ch, ok := c.(*tg.Channel)
			if !ok {
				continue
			}
			if ch.GetID() == p.ChannelID {
				accessHash, _ := ch.GetAccessHash()
				entry := CacheEntry{
					PeerType:   "channel",
					ID:         ch.GetID(),
					AccessHash: accessHash,
					ResolvedAt: nowUnix(),
				}
				return &tg.InputPeerChannel{
					ChannelID:  ch.GetID(),
					AccessHash: accessHash,
				}, entry, nil
			}
		}
		return nil, CacheEntry{}, fmt.Errorf("channel %d not found in resolved response", p.ChannelID)

	default:
		return nil, CacheEntry{}, fmt.Errorf("unsupported peer type: %T", res.Peer)
	}
}

// entryToPeer converts a CacheEntry to an InputPeerClass.
func entryToPeer(e CacheEntry) tg.InputPeerClass {
	switch e.PeerType {
	case "user":
		return &tg.InputPeerUser{UserID: e.ID, AccessHash: e.AccessHash}
	case "channel":
		return &tg.InputPeerChannel{ChannelID: e.ID, AccessHash: e.AccessHash}
	case "chat":
		return &tg.InputPeerChat{ChatID: e.ID}
	default:
		return &tg.InputPeerUser{UserID: e.ID, AccessHash: e.AccessHash}
	}
}

// idToPeer constructs an InputPeerClass from a numeric ID.
//
//   - positive ID        → InputPeerUser{UserID: id}
//   - ID < -1000000000000 → InputPeerChannel{ChannelID: -id - 1000000000000}
//   - other negative     → InputPeerChat{ChatID: -id}
func idToPeer(id int64) tg.InputPeerClass {
	switch {
	case id > 0:
		return &tg.InputPeerUser{UserID: id}
	case id < -1000000000000:
		return &tg.InputPeerChannel{ChannelID: -id - 1000000000000}
	default:
		return &tg.InputPeerChat{ChatID: -id}
	}
}

// nowUnix returns the current time as a Unix timestamp.
func nowUnix() int64 {
	return time.Now().Unix()
}
