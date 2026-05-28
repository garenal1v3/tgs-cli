package search

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/searchtgcli/tgs/internal/search"
)

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
			_, _ = fmt.Fprintf(w, "chat %d:\n", pc.Chat.ID)
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
			_, _ = fmt.Fprintf(w, "chat %d (total=%d):\n", pc.Chat.ID, pc.Total)
			for _, p := range pc.Periods {
				_, _ = fmt.Fprintf(w, "  %s  count=%d  msg=%d..%d\n", p.Date, p.Count, p.MinMsgID, p.MaxMsgID)
			}
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
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
