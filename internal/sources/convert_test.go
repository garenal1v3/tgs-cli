package sources

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestChatToSource_Channel(t *testing.T) {
	ch := &tg.Channel{
		ID:        1234567890,
		Title:     "Durov",
		Broadcast: true,
		Verified:  true,
	}
	ch.SetAccessHash(42)
	ch.SetUsername("durov")

	got := chatToSource(ch)

	if got.Type != "channel" {
		t.Errorf("Type = %q, want channel", got.Type)
	}
	if got.ID != -1001234567890 {
		t.Errorf("ID = %d, want -1001234567890 (Bot API style)", got.ID)
	}
	if got.Title != "Durov" {
		t.Errorf("Title = %q", got.Title)
	}
	if got.Username != "durov" {
		t.Errorf("Username = %q", got.Username)
	}
	if got.Access != "public" {
		t.Errorf("Access = %q, want public", got.Access)
	}
	if !got.Verified {
		t.Error("Verified should be true")
	}
}

func TestChatToSource_Supergroup(t *testing.T) {
	ch := &tg.Channel{
		ID:        555,
		Title:     "Devs",
		Megagroup: true,
		Forum:     true,
		Gigagroup: false,
	}
	ch.SetParticipantsCount(42)

	got := chatToSource(ch)

	if got.Type != "supergroup" {
		t.Errorf("Type = %q, want supergroup", got.Type)
	}
	if got.ID != -1000000000555 {
		t.Errorf("ID = %d, want -1000000000555", got.ID)
	}
	if !got.HasTopics {
		t.Error("HasTopics should be true for forum")
	}
	if got.Access != "private" {
		t.Errorf("Access = %q, want private (no username)", got.Access)
	}
	if got.MembersCount != 42 {
		t.Errorf("MembersCount = %d, want 42", got.MembersCount)
	}
}

func TestChatToSource_LegacyGroup(t *testing.T) {
	c := &tg.Chat{ID: 777, Title: "Legacy", ParticipantsCount: 5}
	got := chatToSource(c)
	if got.Type != "group" {
		t.Errorf("Type = %q, want group", got.Type)
	}
	if got.ID != -777 {
		t.Errorf("ID = %d, want -777", got.ID)
	}
	if got.MembersCount != 5 {
		t.Errorf("MembersCount = %d", got.MembersCount)
	}
}

func TestUserToSource_BotAndUser(t *testing.T) {
	u := &tg.User{ID: 100, Bot: true}
	u.SetFirstName("HelpBot")
	got := userToSource(u, 0)
	if got.Type != "bot" {
		t.Errorf("Type = %q, want bot", got.Type)
	}
	if got.ID != 100 {
		t.Errorf("ID = %d, want 100", got.ID)
	}

	uu := &tg.User{ID: 200}
	uu.SetFirstName("Alice")
	gotUser := userToSource(uu, 0)
	if gotUser.Type != "user" {
		t.Errorf("Type = %q, want user", gotUser.Type)
	}
	if gotUser.FirstName != "Alice" {
		t.Errorf("FirstName = %q", gotUser.FirstName)
	}
}

func TestUserToSource_SelfIsSaved(t *testing.T) {
	u := &tg.User{ID: 999, Self: true}
	got := userToSource(u, 999)
	if !got.Saved {
		t.Error("Saved should be true when user.Self matches self ID")
	}
}

func TestUserToSource_Deleted(t *testing.T) {
	u := &tg.User{ID: 1, Deleted: true}
	got := userToSource(u, 0)
	if !got.Deleted {
		t.Error("Deleted flag should be true")
	}
}

// TestChatToSource_CollectibleUsername covers the case where a channel's
// primary handle has been migrated to the collectible-usernames array
// (Active=true), and the legacy Username field is empty. Without the
// activeUsername fallback such channels reported access="private" with an
// empty username — exactly what we saw for @durov and @money.
func TestChatToSource_CollectibleUsername(t *testing.T) {
	ch := &tg.Channel{ID: 1, Title: "Durov", Broadcast: true}
	ch.SetAccessHash(1)
	// Legacy Username is intentionally NOT set; primary handle lives in
	// the Usernames array as an active entry.
	ch.SetUsernames([]tg.Username{
		{Active: false, Username: "old_handle"},
		{Active: true, Username: "durov"},
	})

	got := chatToSource(ch)

	if got.Username != "durov" {
		t.Errorf("Username = %q, want durov (from collectible Usernames)", got.Username)
	}
	if got.Access != "public" {
		t.Errorf("Access = %q, want public (peer has an active handle)", got.Access)
	}
}

// TestChatToSource_LegacyUsernameWinsOverCollectible verifies that if a peer
// still uses the legacy Username field, we keep using it (Telegram doesn't
// duplicate it into Usernames in that case).
func TestChatToSource_LegacyUsernameWinsOverCollectible(t *testing.T) {
	ch := &tg.Channel{ID: 1, Title: "X", Broadcast: true}
	ch.SetUsername("legacy")
	ch.SetUsernames([]tg.Username{{Active: true, Username: "fallback"}})
	got := chatToSource(ch)
	if got.Username != "legacy" {
		t.Errorf("Username = %q, want legacy (legacy wins when set)", got.Username)
	}
}

// TestChatToSource_NoActiveUsernameStaysPrivate makes sure that a channel
// whose collectible array contains only Active=false entries reports as
// private (we shouldn't pick a disabled handle).
func TestChatToSource_NoActiveUsernameStaysPrivate(t *testing.T) {
	ch := &tg.Channel{ID: 1, Title: "X", Broadcast: true}
	ch.SetUsernames([]tg.Username{{Active: false, Username: "disabled"}})
	got := chatToSource(ch)
	if got.Username != "" {
		t.Errorf("Username = %q, want empty (no active handle)", got.Username)
	}
	if got.Access != "private" {
		t.Errorf("Access = %q, want private", got.Access)
	}
}

// TestUserToSource_CollectibleUsername mirrors the channel test for users:
// the primary handle now lives in user.Usernames with Active=true, and the
// legacy Username field is empty.
func TestUserToSource_CollectibleUsername(t *testing.T) {
	u := &tg.User{ID: 100}
	u.SetFirstName("Pavel")
	u.SetUsernames([]tg.Username{{Active: true, Username: "durov"}})

	got := userToSource(u, 0)

	if got.Username != "durov" {
		t.Errorf("Username = %q, want durov (from collectible Usernames)", got.Username)
	}
}
