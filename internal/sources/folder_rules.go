package sources

import (
	"sort"

	"github.com/gotd/td/tg"
)

// matchFolder applies a DialogFilter / DialogFilterChatlist to a precomputed
// dialog list and returns the folder's resolved contents (a fresh Source per
// match, with Pinned=true for entries that came from pinned_peers).
//
// dialogs is the *complete* set of dialogs we walked (main + optionally
// archive). selfID is the authenticated user's ID — used to resolve
// InputPeerSelf entries in pinned/include lists. Auto-flag rules
// (Task 5) extend this function.
func matchFolder(filter tg.DialogFilterClass, dialogs []sourceWithPeer, selfID int64) []Source {
	index := indexByPeer(dialogs)
	seen := make(map[peerKey]bool)
	var out []Source

	pinned, include := folderPeers(filter)

	for _, p := range pinned {
		k, ok := keyFromInputPeer(p, selfID)
		if !ok {
			continue
		}
		it, found := index[k]
		if !found || seen[k] {
			continue
		}
		src := it.Source
		src.Pinned = true
		out = append(out, src)
		seen[k] = true
	}
	pinnedCount := len(out)

	for _, p := range include {
		k, ok := keyFromInputPeer(p, selfID)
		if !ok {
			continue
		}
		it, found := index[k]
		if !found || seen[k] {
			continue
		}
		src := it.Source
		src.Pinned = false
		out = append(out, src)
		seen[k] = true
	}

	// Step 3: auto-flags — only for DialogFilter, not Chatlist.
	if df, ok := filter.(*tg.DialogFilter); ok {
		excludeSet := buildExcludeSet(df.ExcludePeers, selfID)
		for i := range dialogs {
			it := &dialogs[i]
			k, ok := dialogKey(it)
			if !ok || seen[k] || excludeSet[k] {
				continue
			}
			if df.ExcludeMuted && it.IsMuted {
				continue
			}
			if df.ExcludeRead && it.Source.UnreadCount == 0 {
				continue
			}
			if df.ExcludeArchived && it.Source.Archived {
				continue
			}
			if !matchTypeFlags(df, it) {
				continue
			}
			src := it.Source
			src.Pinned = false
			out = append(out, src)
			seen[k] = true
		}
	}

	// Step 4: sort. Pinned stay in pinned_peers order (first pinnedCount
	// entries); the rest sort by last_message.date descending (nil-date last).
	sortByLastMessage(out[pinnedCount:])

	return out
}

// folderPeers returns the pinned and include peers from a DialogFilterClass,
// abstracting over the three concrete variants.
func folderPeers(filter tg.DialogFilterClass) (pinned, include []tg.InputPeerClass) {
	switch f := filter.(type) {
	case *tg.DialogFilter:
		return f.PinnedPeers, f.IncludePeers
	case *tg.DialogFilterChatlist:
		return f.PinnedPeers, f.IncludePeers
	}
	return nil, nil
}

// peerKey is a comparable identifier for InputPeer/PeerUser/PeerChat/PeerChannel.
// User/chat/channel ID spaces don't collide across kinds.
type peerKey struct {
	kind string // one of peerKindUser / peerKindChat / peerKindChannel
	id   int64
}

const (
	peerKindUser    = "user"
	peerKindChat    = "chat"
	peerKindChannel = "channel"
)

// indexByPeer builds a lookup map from sourceWithPeer.Peer (InputPeer) to the
// dialog snapshot entry. Used by matchFolder to resolve pinned/include refs.
// Note: InputPeerSelf is NOT indexed here — it's expanded at lookup time
// via keyFromInputPeer using the caller-supplied selfID.
func indexByPeer(dialogs []sourceWithPeer) map[peerKey]*sourceWithPeer {
	m := make(map[peerKey]*sourceWithPeer, len(dialogs))
	for i := range dialogs {
		switch p := dialogs[i].Peer.(type) {
		case *tg.InputPeerUser:
			m[peerKey{peerKindUser, p.UserID}] = &dialogs[i]
		case *tg.InputPeerChat:
			m[peerKey{peerKindChat, p.ChatID}] = &dialogs[i]
		case *tg.InputPeerChannel:
			m[peerKey{peerKindChannel, p.ChannelID}] = &dialogs[i]
		}
	}
	return m
}

// keyFromInputPeer returns the peerKey for a folder's pinned/include peer.
// InputPeerSelf resolves to (peerKindUser, selfID).
func keyFromInputPeer(p tg.InputPeerClass, selfID int64) (peerKey, bool) {
	switch v := p.(type) {
	case *tg.InputPeerUser:
		return peerKey{peerKindUser, v.UserID}, true
	case *tg.InputPeerChat:
		return peerKey{peerKindChat, v.ChatID}, true
	case *tg.InputPeerChannel:
		return peerKey{peerKindChannel, v.ChannelID}, true
	case *tg.InputPeerSelf:
		if selfID == 0 {
			return peerKey{}, false
		}
		return peerKey{peerKindUser, selfID}, true
	}
	return peerKey{}, false
}

// buildExcludeSet maps exclude_peers to peerKey for O(1) skip checks.
func buildExcludeSet(peers []tg.InputPeerClass, selfID int64) map[peerKey]bool {
	m := make(map[peerKey]bool, len(peers))
	for _, p := range peers {
		if k, ok := keyFromInputPeer(p, selfID); ok {
			m[k] = true
		}
	}
	return m
}

// dialogKey returns the peerKey for a dialog (built from its InputPeer).
func dialogKey(it *sourceWithPeer) (peerKey, bool) {
	switch p := it.Peer.(type) {
	case *tg.InputPeerUser:
		return peerKey{peerKindUser, p.UserID}, true
	case *tg.InputPeerChat:
		return peerKey{peerKindChat, p.ChatID}, true
	case *tg.InputPeerChannel:
		return peerKey{peerKindChannel, p.ChannelID}, true
	}
	return peerKey{}, false
}

// matchTypeFlags reports whether the dialog matches at least one include-by-
// type flag on the filter. Bot/non-contact distinction: a bot user does NOT
// count as non_contacts (per Telegram client behaviour).
func matchTypeFlags(df *tg.DialogFilter, it *sourceWithPeer) bool {
	t := it.Source.Type
	if df.Contacts && (t == "user" || t == "bot") && it.IsContact {
		return true
	}
	if df.NonContacts && t == "user" && !it.IsContact {
		return true
	}
	if df.Groups && (t == "group" || t == "supergroup") {
		return true
	}
	if df.Broadcasts && t == "channel" {
		return true
	}
	if df.Bots && t == "bot" {
		return true
	}
	return false
}

// sortByLastMessage sorts in place by LastMessage.Date desc; entries with nil
// LastMessage go to the end. Stable so equal dates keep their original order.
func sortByLastMessage(src []Source) {
	sort.SliceStable(src, func(i, j int) bool {
		a, b := src[i].LastMessage, src[j].LastMessage
		switch {
		case a == nil && b == nil:
			return false
		case a == nil:
			return false
		case b == nil:
			return true
		default:
			return a.Date > b.Date // ISO-8601 strings sort lexicographically
		}
	})
}
