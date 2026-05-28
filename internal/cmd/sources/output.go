package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	sourcessvc "github.com/searchtgcli/tgs/internal/sources"
)

func writeSourceList(w io.Writer, format string, result *sourcessvc.ListResult) error {
	if format == "text" {
		for _, s := range result.Sources {
			writeSourceLine(w, s)
		}
		if result.Cursor != "" {
			_, _ = fmt.Fprintf(w, "\n--- cursor: %s\n", result.Cursor)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

func writeSource(w io.Writer, format string, src *sourcessvc.Source) error {
	if format == "text" {
		writeSourceKV(w, src)
		return nil
	}
	return json.NewEncoder(w).Encode(src)
}

// writeSourceLine renders one source as a single tabular row.
func writeSourceLine(w io.Writer, s sourcessvc.Source) {
	username := ""
	if s.Username != "" {
		username = "@" + s.Username
	}
	last := ""
	if s.LastMessage != nil {
		last = s.LastMessage.Date
		if len(last) > 19 {
			last = last[:19]
		}
	}
	cols := []string{
		fmt.Sprintf("%-10s", s.Type),
		fmt.Sprintf("%14d", s.ID),
		fmt.Sprintf("%-30s", trunc(displayTitle(s), 30)),
		fmt.Sprintf("%-20s", username),
		fmt.Sprintf("u=%-4d", s.UnreadCount),
		last,
	}
	if s.Stats != nil {
		cols = append(cols,
			fmt.Sprintf("m=%-7d", s.MembersCount),
			fmt.Sprintf("total=%-7d", s.Stats.TotalMessages),
			fmt.Sprintf("24h=%-4d", s.Stats.Messages24h),
		)
		if s.Stats.FirstMessage != nil {
			cols = append(cols, "first="+s.Stats.FirstMessage.Date[:10])
		}
	}
	_, _ = fmt.Fprintln(w, strings.Join(cols, "  "))
}

// writeSourceKV renders one source as aligned key:value pairs.
//
// The text format is a flat projection of every populated Source field — if
// the JSON has it, the text view should too. Fields are emitted only when
// non-zero so that callers can grep cleanly. Order is biased toward identity
// first (id/type/title/username), then access/membership, then person-only
// fields (first/last name, phone), then flags, then full-info, then stats.
func writeSourceKV(w io.Writer, s *sourcessvc.Source) {
	line := func(k string, v interface{}) {
		_, _ = fmt.Fprintf(w, "%-17s %v\n", k+":", v)
	}
	yes := func(k string, b bool) {
		if b {
			line(k, true)
		}
	}

	line("id", s.ID)
	line("type", s.Type)
	if s.Title != "" {
		line("title", s.Title)
	}
	if s.Username != "" {
		line("username", "@"+s.Username)
	}
	if s.Access != "" {
		line("access", s.Access)
	}
	if s.MembersCount > 0 {
		line("members", s.MembersCount)
	}

	if s.FirstName != "" {
		line("first_name", s.FirstName)
	}
	if s.LastName != "" {
		line("last_name", s.LastName)
	}
	if s.Phone != "" {
		line("phone", "+"+s.Phone)
	}

	yes("verified", s.Verified)
	yes("scam", s.Scam)
	yes("fake", s.Fake)
	yes("restricted", s.Restricted)
	if s.RestrictedReason != "" {
		line("restricted_reason", s.RestrictedReason)
	}
	yes("deleted", s.Deleted)
	yes("archived", s.Archived)
	yes("pinned", s.Pinned)
	yes("saved", s.Saved)
	yes("has_topics", s.HasTopics)
	yes("gigagroup", s.Gigagroup)
	yes("has_comments", s.HasComments)
	if s.LinkedChatID != 0 {
		line("linked_chat_id", s.LinkedChatID)
	}

	line("unread_count", s.UnreadCount)
	if s.LastMessage != nil {
		line("last_msg", fmt.Sprintf("#%d at %s", s.LastMessage.ID, s.LastMessage.Date))
	}
	if s.Subscribed != nil {
		line("subscribed", *s.Subscribed)
	}
	if s.Description != "" {
		line("description", s.Description)
	}
	if s.CreationDate != "" {
		line("creation_date", s.CreationDate)
	}
	if s.InviteLink != "" {
		line("invite_link", s.InviteLink)
	}
	if s.Stats != nil {
		line("total_msgs", s.Stats.TotalMessages)
		line("msgs_24h", s.Stats.Messages24h)
		if s.Stats.FirstMessage != nil {
			line("first_msg", fmt.Sprintf("#%d at %s", s.Stats.FirstMessage.ID, s.Stats.FirstMessage.Date))
		}
	}
	if s.StatsError != "" {
		line("stats_error", s.StatsError)
	}
}

// displayTitle returns a non-empty label for the title column. Channels and
// groups use Title verbatim; for users/bots — where Title is empty by design —
// it falls back to "first_name last_name" so the row isn't a hole in the
// table.
func displayTitle(s sourcessvc.Source) string {
	if s.Title != "" {
		return s.Title
	}
	name := strings.TrimSpace(s.FirstName + " " + s.LastName)
	return name
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func writeFoldersResult(w io.Writer, format string, result *sourcessvc.FoldersResult) error {
	if format == "text" {
		_, _ = fmt.Fprintf(w, "%d folders\n", result.Total)
		for _, f := range result.Folders {
			_, _ = fmt.Fprintln(w)
			emoticon := ""
			if f.Emoticon != "" {
				emoticon = " " + f.Emoticon
			}
			_, _ = fmt.Fprintf(w, "[%s] %s%s (id=%d) — %d chats\n", f.Kind, f.Title, emoticon, f.ID, f.ChatsCount)
			for _, s := range f.Chats {
				writeFolderChatLine(w, s)
			}
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

// writeFolderChatLine renders one chat row under a folder in the text view.
func writeFolderChatLine(w io.Writer, s sourcessvc.Source) {
	username := ""
	if s.Username != "" {
		username = "@" + s.Username
	}
	flags := ""
	if s.UnreadCount > 0 {
		flags += fmt.Sprintf("  unread=%d", s.UnreadCount)
	}
	if s.Pinned {
		flags += "  pinned"
	}
	_, _ = fmt.Fprintf(w, "  %-10s %s %s%s\n", s.Type, trunc(displayTitle(s), 30), username, flags)
}
