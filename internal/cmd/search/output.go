package search

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/searchtgcli/tgs/internal/search"
	"github.com/searchtgcli/tgs/internal/sources"
)

// chatRefFromSource builds a ChatRef from a resolved folder source. Source.ID
// is already in Bot-API form, and Type/Title/Username carry the display data
// fan-out results expose per chat.
func chatRefFromSource(s sources.Source) search.ChatRef {
	return search.ChatRef{ID: s.ID, Type: s.Type, Title: s.Title, Username: s.Username}
}

// chatRefFromPeer builds a ChatRef from a bare InputPeer (a --chat value with
// no resolved metadata). ID is converted to Bot-API form; Type is only known
// for legacy groups (InputPeerChat), and Title/Username stay empty.
func chatRefFromPeer(p tg.InputPeerClass) search.ChatRef {
	switch v := p.(type) {
	case *tg.InputPeerChannel:
		return search.ChatRef{ID: -1000000000000 - v.ChannelID}
	case *tg.InputPeerChat:
		return search.ChatRef{ID: -v.ChatID, Type: "group"}
	case *tg.InputPeerUser:
		return search.ChatRef{ID: v.UserID}
	}
	return search.ChatRef{}
}

func outputFormat(cmd *cobra.Command) string {
	if f := cmd.Flag("output"); f != nil {
		return f.Value.String()
	}
	return "json"
}

func writeSearchResult(w io.Writer, format string, result *search.SearchResult) error {
	if format == "text" {
		for _, m := range result.Messages {
			chatName := m.Chat.Title
			if chatName == "" && m.Chat.Username != "" {
				chatName = "@" + m.Chat.Username
			}
			if chatName == "" {
				chatName = fmt.Sprintf("#%d", m.Chat.ID)
			}

			from := ""
			if m.From != nil {
				if m.From.Username != "" {
					from = "@" + m.From.Username
				} else {
					parts := []string{}
					if m.From.FirstName != "" {
						parts = append(parts, m.From.FirstName)
					}
					if m.From.LastName != "" {
						parts = append(parts, m.From.LastName)
					}
					if len(parts) > 0 {
						from = strings.Join(parts, " ")
					} else {
						from = fmt.Sprintf("#%d", m.From.ID)
					}
				}
			}

			date := m.Date
			if len(date) > 19 {
				date = date[:19]
			}
			date = strings.Replace(date, "T", " ", 1)

			text := m.Text
			if len(text) > 200 {
				text = text[:200] + "..."
			}
			text = strings.ReplaceAll(text, "\n", " ")

			stats := []string{}
			if m.Views > 0 {
				stats = append(stats, fmt.Sprintf("👁 %d", m.Views))
			}
			if m.Forwards > 0 {
				stats = append(stats, fmt.Sprintf("↻ %d", m.Forwards))
			}
			if m.Replies > 0 {
				stats = append(stats, fmt.Sprintf("💬 %d", m.Replies))
			}
			for _, r := range m.Reactions {
				stats = append(stats, fmt.Sprintf("%s %d", r.Emoji, r.Count))
			}
			statsStr := ""
			if len(stats) > 0 {
				statsStr = " [" + strings.Join(stats, " ") + "]"
			}

			if from != "" {
				_, _ = fmt.Fprintf(w, "[%s] %s | %s: %s%s\n", date, chatName, from, text, statsStr)
			} else {
				_, _ = fmt.Fprintf(w, "[%s] %s: %s%s\n", date, chatName, text, statsStr)
			}
		}
		if result.Cursor != "" {
			_, _ = fmt.Fprintf(w, "\n--- cursor: %s\n", result.Cursor)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

func writeCountersResult(w io.Writer, format string, result *search.CountersResult) error {
	if format == "text" {
		for _, c := range result.Counters {
			_, _ = fmt.Fprintf(w, "%-12s %d\n", c.Filter, c.Count)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

func writeMultiCountersResult(w io.Writer, format string, result *search.MultiCountersResult) error {
	if format == "text" {
		for _, pc := range result.Chats {
			_, _ = fmt.Fprintf(w, "%s:\n", chatLabel(pc.Chat))
			for _, c := range pc.Counters {
				_, _ = fmt.Fprintf(w, "  %-10s %d\n", c.Filter, c.Count)
			}
		}
		_, _ = fmt.Fprintln(w, "totals:")
		for _, c := range result.Totals {
			_, _ = fmt.Fprintf(w, "  %-10s %d\n", c.Filter, c.Count)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

func writeMultiCalendarResult(w io.Writer, format string, result *search.MultiCalendarResult) error {
	if format == "text" {
		for _, pc := range result.Chats {
			_, _ = fmt.Fprintf(w, "%s (total=%d):\n", chatLabel(pc.Chat), pc.Total)
			for _, p := range pc.Periods {
				_, _ = fmt.Fprintf(w, "  %s  count=%d  msg=%d..%d\n", p.Date, p.Count, p.MinMsgID, p.MaxMsgID)
			}
		}
		if len(result.Totals) > 0 {
			_, _ = fmt.Fprintln(w, "totals:")
			for _, t := range result.Totals {
				_, _ = fmt.Fprintf(w, "  %s  %d\n", t.Date, t.Count)
			}
			_, _ = fmt.Fprintf(w, "\ntotal: %d\n", result.Total)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

// chatLabel renders a ChatRef for text output: prefer title, then @username,
// then the numeric id.
func chatLabel(c search.ChatRef) string {
	if c.Title != "" {
		return c.Title
	}
	if c.Username != "" {
		return "@" + c.Username
	}
	return fmt.Sprintf("#%d", c.ID)
}

func writeCalendarResult(w io.Writer, format string, result *search.CalendarResult) error {
	if format == "text" {
		for _, p := range result.Periods {
			_, _ = fmt.Fprintf(w, "%s  %d\n", p.Date, p.Count)
		}
		if result.Total > 0 {
			_, _ = fmt.Fprintf(w, "\ntotal: %d\n", result.Total)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}
