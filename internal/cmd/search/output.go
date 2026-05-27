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

			if from != "" {
				fmt.Fprintf(w, "[%s] %s | %s: %s\n", date, chatName, from, text)
			} else {
				fmt.Fprintf(w, "[%s] %s: %s\n", date, chatName, text)
			}
		}
		if result.Cursor != "" {
			fmt.Fprintf(w, "\n--- cursor: %s\n", result.Cursor)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

func writeCountersResult(w io.Writer, format string, result *search.CountersResult) error {
	if format == "text" {
		for _, c := range result.Counters {
			fmt.Fprintf(w, "%-12s %d\n", c.Filter, c.Count)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}

func writeCalendarResult(w io.Writer, format string, result *search.CalendarResult) error {
	if format == "text" {
		for _, p := range result.Periods {
			fmt.Fprintf(w, "%s  %d\n", p.Date, p.Count)
		}
		if result.Total > 0 {
			fmt.Fprintf(w, "\ntotal: %d\n", result.Total)
		}
		return nil
	}
	return json.NewEncoder(w).Encode(result)
}
