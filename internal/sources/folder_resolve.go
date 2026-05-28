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
func (s *Service) ResolveFolder(ctx context.Context, ref string, archived bool) (*ResolvedFolder, error) {
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

	dialogs, err := s.fetchAllDialogs(ctx, archived)
	if err != nil {
		return nil, err
	}

	folder, skip := buildFolder(filter, dialogs, s.selfID)
	if skip {
		// Should never happen — findFolderFilter filters out Default — but
		// guard anyway.
		return nil, fmt.Errorf("folder not found: %q", ref)
	}

	// Collect InputPeers for fan-out, in the same order as Folder.Chats.
	peers := make([]tg.InputPeerClass, 0, len(folder.Chats))
	index := indexByPeer(dialogs)
	for _, src := range folder.Chats {
		k := sourceKey(src)
		if it, ok := index[k]; ok && it.Peer != nil {
			peers = append(peers, it.Peer)
		}
	}

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
// indexByPeer. Source.ID uses Bot-API conventions: negative for channels
// (-100...), negative for legacy chats, positive for users.
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
