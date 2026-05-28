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

func TestMatchFolder_ContactsFlag(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(1, "user", "Alice", func(w *sourceWithPeer) { w.IsContact = true }),
		swp(2, "user", "Bob", func(w *sourceWithPeer) { w.IsContact = false }),
		swp(3, "bot", "Botty"),
	}
	filter := &tg.DialogFilter{ID: 1, Title: tg.TextWithEntities{Text: "Contacts"}, Contacts: true}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != 1 {
		t.Errorf("got = %+v, want only Alice", got)
	}
}

func TestMatchFolder_NonContactsFlag(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(1, "user", "Alice", func(w *sourceWithPeer) { w.IsContact = true }),
		swp(2, "user", "Bob"),
		swp(3, "bot", "Botty"),
	}
	filter := &tg.DialogFilter{ID: 1, Title: tg.TextWithEntities{Text: "Non-contacts"}, NonContacts: true}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != 2 {
		t.Errorf("got = %+v, want only Bob", got)
	}
}

func TestMatchFolder_GroupsBroadcastsBotsFlags(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "Ch"),
		swp(-1000000000002, "supergroup", "SGrp"),
		swp(-3, "group", "LegacyGrp"),
		swp(4, "user", "Alice"),
		swp(5, "bot", "Botty"),
	}
	filter := &tg.DialogFilter{ID: 1, Title: tg.TextWithEntities{Text: "GBB"}, Groups: true, Broadcasts: true, Bots: true}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4 (channel+sgroup+group+bot)", len(got))
	}
}

func TestMatchFolder_ExcludePeersBeatsAutoInclude(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "A"),
		swp(-1000000000002, "channel", "B"),
	}
	filter := &tg.DialogFilter{
		ID:           1,
		Title:        tg.TextWithEntities{Text: "Channels"},
		Broadcasts:   true,
		ExcludePeers: []tg.InputPeerClass{&tg.InputPeerChannel{ChannelID: 2, AccessHash: 1}},
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != -1000000000001 {
		t.Errorf("got = %+v, want only A", got)
	}
}

func TestMatchFolder_ExcludeMutedReadArchived(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "Active"),
		swp(-1000000000002, "channel", "Muted", func(w *sourceWithPeer) { w.IsMuted = true }),
		swp(-1000000000003, "channel", "Read", func(w *sourceWithPeer) { w.Source.UnreadCount = 0 }),
		swp(-1000000000004, "channel", "Arch", func(w *sourceWithPeer) { w.Source.Archived = true }),
	}
	// Give the "Active" one unread so ExcludeRead doesn't eat it.
	dialogs[0].Source.UnreadCount = 1
	filter := &tg.DialogFilter{
		ID:              1,
		Title:           tg.TextWithEntities{Text: "Strict"},
		Broadcasts:      true,
		ExcludeMuted:    true,
		ExcludeRead:     true,
		ExcludeArchived: true,
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != -1000000000001 {
		t.Errorf("got = %+v, want only Active", got)
	}
}

func TestMatchFolder_ChatlistIgnoresAutoFlags(t *testing.T) {
	// A DialogFilterChatlist with include_peers — chatlists have no
	// type-include flags, so dialogs not in include_peers must NOT appear.
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "A"),
		swp(-1000000000002, "channel", "B"),
	}
	filter := &tg.DialogFilterChatlist{
		ID:    1,
		Title: tg.TextWithEntities{Text: "Shared"},
		IncludePeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1},
		},
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 1 || got[0].ID != -1000000000001 {
		t.Errorf("got = %+v, want only A (B must NOT auto-include)", got)
	}
}

func TestMatchFolder_SortPinnedFirstThenByLastMsgDesc(t *testing.T) {
	dialogs := []sourceWithPeer{
		swp(-1000000000001, "channel", "Oldest", func(w *sourceWithPeer) {
			w.Source.LastMessage = &LastMessage{ID: 1, Date: "2026-05-20T10:00:00Z"}
		}),
		swp(-1000000000002, "channel", "Newest", func(w *sourceWithPeer) {
			w.Source.LastMessage = &LastMessage{ID: 2, Date: "2026-05-25T10:00:00Z"}
		}),
		swp(-1000000000003, "channel", "Pinned", func(w *sourceWithPeer) {
			w.Source.LastMessage = &LastMessage{ID: 3, Date: "2026-05-15T10:00:00Z"}
		}),
	}
	filter := &tg.DialogFilter{
		ID:    1,
		Title: tg.TextWithEntities{Text: "T"},
		PinnedPeers: []tg.InputPeerClass{
			&tg.InputPeerChannel{ChannelID: 3, AccessHash: 1},
		},
		Broadcasts: true,
	}
	got := matchFolder(filter, dialogs, 0)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].ID != -1000000000003 {
		t.Errorf("got[0].ID = %d, want -1000000000003 (Pinned)", got[0].ID)
	}
	if got[1].ID != -1000000000002 {
		t.Errorf("got[1].ID = %d, want -1000000000002 (Newest)", got[1].ID)
	}
	if got[2].ID != -1000000000001 {
		t.Errorf("got[2].ID = %d, want -1000000000001 (Oldest)", got[2].ID)
	}
}
