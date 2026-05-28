package sources

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	sourcessvc "github.com/searchtgcli/tgs/internal/sources"
)

func TestTrunc_HandlesCyrillic(t *testing.T) {
	got := trunc("ПриветМирЭтоДлинноеНазвание", 10)
	if utf8.RuneCountInString(got) != 10 {
		t.Errorf("expected 10 runes, got %d in %q", utf8.RuneCountInString(got), got)
	}
	if !utf8.ValidString(got) {
		t.Errorf("output is not valid UTF-8: %q", got)
	}
}

func TestWriteSource_TextHasKeyLines(t *testing.T) {
	src := &sourcessvc.Source{
		ID: -1001234567890, Type: "channel", Title: "Durov",
		Username: "durov", Access: "public",
		UnreadCount: 0,
	}
	var buf bytes.Buffer
	if err := writeSource(&buf, "text", src); err != nil {
		t.Fatalf("writeSource: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"id:", "type:", "title:", "username:", "Durov", "durov", "channel"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s\n---", want, out)
		}
	}
}

// TestWriteSource_TextShowsExtendedFields covers the fields that were missing
// from the text view before bug #6 was fixed: verified flag, creation_date,
// invite_link, has_comments + linked_chat_id, person fields, and the
// optional boolean flags.
func TestWriteSource_TextShowsExtendedFields(t *testing.T) {
	yes := true
	src := &sourcessvc.Source{
		ID: -1001234567890, Type: "channel", Title: "Pavel Durov",
		Username: "durov", Access: "public",
		MembersCount: 1234567,
		Verified:     true,
		HasComments:  true,
		LinkedChatID: -1009876543210,
		Subscribed:   &yes,
		Description:  "Founder of Telegram.",
		CreationDate: "2015-08-26T10:00:00Z",
		InviteLink:   "https://t.me/+abc",
		Stats: &sourcessvc.Stats{
			TotalMessages: 12345,
			Messages24h:   3,
			FirstMessage:  &sourcessvc.FirstMessage{ID: 1, Date: "2015-08-26T10:00:00Z"},
		},
	}
	var buf bytes.Buffer
	if err := writeSource(&buf, "text", src); err != nil {
		t.Fatalf("writeSource: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"verified:", "creation_date:", "2015-08-26T10:00:00Z",
		"invite_link:", "https://t.me/+abc",
		"has_comments:", "linked_chat_id:", "-1009876543210",
		"description:", "Founder of Telegram.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q\n---\n%s\n---", want, out)
		}
	}
}

// TestWriteSource_TextUserFields verifies user-only fields land in the text
// output for a bot/user (first_name, last_name, phone).
func TestWriteSource_TextUserFields(t *testing.T) {
	src := &sourcessvc.Source{
		ID: 100, Type: "user",
		FirstName: "Alice", LastName: "Smith", Phone: "79001234567",
	}
	var buf bytes.Buffer
	if err := writeSource(&buf, "text", src); err != nil {
		t.Fatalf("writeSource: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"first_name:", "Alice", "last_name:", "Smith", "phone:", "+79001234567"} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q\n---\n%s\n---", want, out)
		}
	}
}

// TestWriteSourceList_TextFallsBackToFirstNameForUser verifies that the
// tabular `list` text output uses first_name (+ last_name) for users and bots
// — which have no Title field — instead of leaving the column blank.
func TestWriteSourceList_TextFallsBackToFirstNameForUser(t *testing.T) {
	result := &sourcessvc.ListResult{
		Sources: []sourcessvc.Source{
			{ID: 100, Type: "bot", Username: "helpbot", FirstName: "Help"},
			{ID: 200, Type: "user", FirstName: "Alice", LastName: "Smith"},
		},
	}
	var buf bytes.Buffer
	if err := writeSourceList(&buf, "text", result); err != nil {
		t.Fatalf("writeSourceList: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Help") {
		t.Errorf("bot row missing first_name fallback (want 'Help'):\n%s", out)
	}
	if !strings.Contains(out, "Alice Smith") {
		t.Errorf("user row missing first_name+last_name fallback:\n%s", out)
	}
}

func TestWriteSourceList_JSON(t *testing.T) {
	result := &sourcessvc.ListResult{
		Sources: []sourcessvc.Source{{ID: 1, Type: "user", Title: "Alice"}},
		Total:   1,
	}
	var buf bytes.Buffer
	if err := writeSourceList(&buf, "json", result); err != nil {
		t.Fatalf("writeSourceList: %v", err)
	}
	if !strings.Contains(buf.String(), `"id":1`) {
		t.Errorf("expected id:1 in JSON, got: %s", buf.String())
	}
}
