package sources

import (
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

	// Step 1: pinned (in order).
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

	// Step 2: include.
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
		// Don't override Source.Pinned from the dialog snapshot — leave it as
		// it came; pinned-in-folder semantics only apply for the pinned_peers
		// list. But the snapshot's pinned flag (pinned in main view) is
		// unrelated to folder UI, so override to false here for consistency.
		src.Pinned = false
		out = append(out, src)
		seen[k] = true
	}

	// Step 3: auto-flags — added in Task 5.

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
