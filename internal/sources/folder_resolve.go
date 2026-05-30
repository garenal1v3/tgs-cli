package sources

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
)

// ResolveFolder finds a user-defined folder by either its numeric ID or its
// title (case-insensitive exact match), then materialises its contents and
// returns the matching InputPeers alongside.
//
// DialogFilterDefault ("All chats") is intentionally NOT resolvable — it has
// no rules and represents the absence of a folder filter.
//
// The folder's contents always include archived chats (see fetchAllDialogs):
// a folder is a view that can span the archive, so scoping a command to a
// folder must surface every chat in it regardless of archive state.
func (s *Service) ResolveFolder(ctx context.Context, ref string) (*ResolvedFolder, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, errors.New("folder ref is empty")
	}

	filters, err := s.fetchDialogFilters(ctx)
	if err != nil {
		return nil, err
	}

	filter, err := findFolderFilter(filters, ref)
	if err != nil {
		return nil, err
	}

	dialogs, err := s.fetchAllDialogs(ctx)
	if err != nil {
		return nil, err
	}

	folder, skip := buildFolder(filter, dialogs, s.selfID)
	if skip {
		// Should never happen — findFolderFilter filters out Default — but
		// guard anyway.
		return nil, fmt.Errorf("folder not found: %q", ref)
	}

	// Build Chats and their InputPeers in lockstep so callers can rely on the
	// invariant folder.Chats[i] <-> peers[i] (fan-out commands pair each peer's
	// result with the chat's display metadata). A chat without a resolvable
	// InputPeer is dropped from BOTH — it can't be queried anyway, and keeping
	// it would desync the two slices. In practice every chat sourced from the
	// dialog list has a peer, so nothing is dropped.
	index := indexByPeer(dialogs)
	chats := make([]Source, 0, len(folder.Chats))
	peers := make([]tg.InputPeerClass, 0, len(folder.Chats))
	for _, src := range folder.Chats {
		k := sourceKey(src)
		it, ok := index[k]
		if !ok || it.Peer == nil {
			continue
		}
		chats = append(chats, src)
		peers = append(peers, it.Peer)
	}
	folder.Chats = chats
	folder.ChatsCount = len(chats)

	return &ResolvedFolder{Folder: folder, Peers: peers}, nil
}

// findFolderFilter locates a non-default folder filter by id or title.
func findFolderFilter(filters []tg.DialogFilterClass, ref string) (tg.DialogFilterClass, error) {
	// Numeric? Try ID match first (but skip Default which has no ID accessor
	// — Default never has user-visible ID).
	if id, err := strconv.Atoi(ref); err == nil {
		for _, f := range filters {
			switch v := f.(type) {
			case *tg.DialogFilter:
				if v.ID == id {
					return f, nil
				}
			case *tg.DialogFilterChatlist:
				if v.ID == id {
					return f, nil
				}
			}
		}
		return nil, fmt.Errorf("folder not found: %q", ref)
	}

	// Title match (case-insensitive exact).
	var matches []tg.DialogFilterClass
	var matchedIDs []int
	for _, f := range filters {
		switch v := f.(type) {
		case *tg.DialogFilter:
			if strings.EqualFold(v.Title.Text, ref) {
				matches = append(matches, f)
				matchedIDs = append(matchedIDs, v.ID)
			}
		case *tg.DialogFilterChatlist:
			if strings.EqualFold(v.Title.Text, ref) {
				matches = append(matches, f)
				matchedIDs = append(matchedIDs, v.ID)
			}
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("folder not found: %q", ref)
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("ambiguous folder name %q (matches ids %v)", ref, matchedIDs)
	}
}

// sourceKey returns the peerKey for a Source so it can be looked up against
// indexByPeer. Source.ID uses Bot-API conventions:
//   - channels: s.ID = -1_000_000_000_000 - tg_channel_id, inverted here.
//   - legacy groups: s.ID = -tg_chat_id, so -s.ID recovers chat_id.
//   - users/bots: s.ID = tg_user_id (identity).
func sourceKey(s Source) peerKey {
	switch s.Type {
	case "channel", "supergroup":
		return peerKey{peerKindChannel, -s.ID - 1000000000000}
	case "group":
		return peerKey{peerKindChat, -s.ID}
	case "user", "bot":
		return peerKey{peerKindUser, s.ID}
	}
	return peerKey{}
}
