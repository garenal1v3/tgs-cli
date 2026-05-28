package sources

import (
	"testing"

	"github.com/gotd/td/tg"
)

// helper: build a sourceWithPeer for tests.
func swp(id int64, typ, title string, opts ...func(*sourceWithPeer)) sourceWithPeer {
	w := sourceWithPeer{
		Source: Source{ID: id, Type: typ, Title: title},
	}
	switch typ {
	case "channel", "supergroup":
		w.Peer = &tg.InputPeerChannel{ChannelID: channelIDFromBotAPI(id), AccessHash: 1}
	case "group":
		w.Peer = &tg.InputPeerChat{ChatID: -id}
	case "user", "bot":
		w.Peer = &tg.InputPeerUser{UserID: id, AccessHash: 1}
	}
	for _, o := range opts {
		o(&w)
	}
	return w
}

// channelIDFromBotAPI converts -1001234567890 back to 1234567890 for the
// InputPeerChannel.ChannelID field — inverse of botAPIChannelID.
func channelIDFromBotAPI(id int64) int64 {
	if id < 0 {
		return -id - 1000000000000
	}
	return id
}

func TestMatchFolder_PinnedFirstInOrder(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "A"),
		swp(-1000000000002, "channel", "B"),
		swp(-1000000000003, "channel", "C"),
	}
	filter := &tg.DialogFilter{
		ID:    1,
		Title: tg.TextWithEntities{Text: "Test"},
		PinnedPeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 2, AccessHash: 1},
			&tg.InputPeerChannel{ChannelID: 3, AccessHash: 1},
		},
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != -1000000000002 || !got[0].Pinned {
		t.Errorf("got[0] = %+v, want id=-1000000000002 pinned=true", got[0])
	}
	if got[1].ID != -1000000000003 || !got[1].Pinned {
		t.Errorf("got[1] = %+v, want id=-1000000000003 pinned=true", got[1])
	}
}

func TestMatchFolder_IncludeAddedNotPinned(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "A"),
		swp(-1000000000002, "channel", "B"),
	}
	filter := &tg.DialogFilter{
		ID:    1,
		Title: tg.TextWithEntities{Text: "Test"},
		IncludePeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
		},
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != -1000000000001 {
		t.Fatalf("got = %+v", got)
	}
	if got[0].Pinned {
		t.Errorf("Pinned = true, want false (include-only peer)")
	}
}

func TestMatchFolder_PinnedAndIncludeDeduped(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "A"),
	}
	filter := &tg.DialogFilter{
		ID:    1,
		Title: tg.TextWithEntities{Text: "Test"},
		PinnedPeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
		},
		IncludePeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
		},
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 {
		t.Errorf("len = %d, want 1 (deduped)", len(got))
	}
	if !got[0].Pinned {
		t.Errorf("Pinned = false, want true (pinned wins)")
	}
}

func TestMatchFolder_PinnedPeerSelfResolvesBySelfID(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(777, "user", "Me"),
	}
	filter := &tg.DialogFilter{
		ID:    1,
		Title: tg.TextWithEntities{Text: "Test"},
		PinnedPeers: []tg.InputPeerClass{
			&tg.InputPeerSelf{},
		},
	}
	got := matchFolder(filter, dialogs, 777)
	if len(got) != 1 || got[0].ID != 777 {
		t.Fatalf("got = %+v", got)
	}
	if !got[0].Pinned {
		t.Errorf("got[0].Pinned = false, want true (came via PinnedPeers)")
	}
}

func TestMatchFolder_SelfWithZeroSelfIDSkipped(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(777, "user", "Me"),
	}
	filter := &tg.DialogFilter{
		ID:          1,
		Title:       tg.TextWithEntities{Text: "T"},
		PinnedPeers: []tg.InputPeerClass{&tg.InputPeerSelf{}},
	}
	got := matchFolder(filter, dialogs, 0) // selfID=0 means unknown
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0 (Self with unknown selfID must be skipped)", len(got))
	}
}

func TestMatchFolder_ChatlistInclude(t *testing.T) {
	dialogs := []sourceWithPeer{swp(-1000000000001, "channel", "A")}
	filter := &tg.DialogFilterChatlist{
		ID:    1,
		Title: tg.TextWithEntities{Text: "Shared"},
		IncludePeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
		},
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != -1000000000001 {
		t.Errorf("got = %+v", got)
	}
}

func TestMatchFolder_PeerNotInDialogsSilentlySkipped(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "A"),
	}
	filter := &tg.DialogFilter{
		ID:    1,
		Title: tg.TextWithEntities{Text: "Test"},
		PinnedPeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 999, AccessHash: 1}, // not in dialogs
		},
		IncludePeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
		},
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != -1000000000001 {
		t.Errorf("got = %+v, want only -1000000000001", got)
	}
}
