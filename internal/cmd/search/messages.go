package search

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/searchtgcli/tgs/internal/profile"
	"github.com/searchtgcli/tgs/internal/resolve"
	"github.com/searchtgcli/tgs/internal/retry"
	"github.com/searchtgcli/tgs/internal/search"
	"github.com/searchtgcli/tgs/internal/sources"
	"github.com/searchtgcli/tgs/internal/telegram"
)

func newMessagesCmd() *cobra.Command {
	var (
		flagChat            []string
		flagFolder          string
		flagFrom            string
		flagFilter          string
		flagAfter           string
		flagBefore          string
		flagTopic           int
		flagLimit           int
		flagCursor          string
		flagMaxWait         int
		flagNoCache         bool
		flagProfile         string
		flagIncludeComments bool
	)

	cmd := &cobra.Command{
		Use:   "messages [query]",
		Short: "Search messages in specific chats",
		Long:  "Search messages within one or more chats, channels, or groups.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			chats := expandChats(flagChat)
			if len(chats) == 0 && flagFolder == "" {
				return fmt.Errorf("at least one of --chat/-c or --folder is required")
			}

			filter, err := search.ParseFilter(flagFilter)
			if err != nil {
				return err
			}

			after, err := parseDate(flagAfter)
			if err != nil {
				return fmt.Errorf("invalid --after: %w", err)
			}
			before, err := parseDate(flagBefore)
			if err != nil {
				return fmt.Errorf("invalid --before: %w", err)
			}

			if err := validateLimit(flagLimit); err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			profileName := profile.Resolve(flagProfile, cwd)

			client, err := telegram.OpenOrError(profileName)
			if err != nil {
				return err
			}
			defer func() { _ = client.Close() }()

			ctx := cmd.Context()
			return client.Run(ctx, func(ctx context.Context, api *tg.Client) error {
				retryPolicy := retry.DefaultPolicy()
				retryPolicy.MaxFloodWait = time.Duration(flagMaxWait) * time.Second

				var cache *resolve.PeerCache
				if !flagNoCache {
					cachePath := telegram.CachePath(profileName)
					c, err := resolve.NewPeerCache(cachePath)
					if err != nil {
						return fmt.Errorf("open peer cache: %w", err)
					}
					defer func() { _ = c.Close() }()
					cache = c
				}

				resolver := resolve.NewResolver(api, cache)

				var peers []tg.InputPeerClass
				if len(chats) > 0 {
					var err error
					peers, err = resolver.ResolveMulti(ctx, chats)
					if err != nil {
						return fmt.Errorf("resolve chats: %w", err)
					}
				}

				if flagFolder != "" {
					meta, _ := client.LoadMeta()
					var selfID int64
					if meta != nil {
						selfID = meta.ID
					}
					src := sources.New(api, resolver, retryPolicy, nil, selfID)
					resolved, err := src.ResolveFolder(ctx, flagFolder, false)
					if err != nil {
						return fmt.Errorf("resolve --folder: %w", err)
					}
					peers = append(peers, resolved.Peers...)
				}

				if len(peers) == 0 {
					return fmt.Errorf("no chats to search (--folder %q is empty and no --chat given)", flagFolder)
				}

				if flagIncludeComments {
					linked := resolveLinkedChats(ctx, api, peers)
					peers = append(peers, linked...)
				}

				var fromPeer tg.InputPeerClass
				if flagFrom != "" {
					fromPeer, err = resolver.Resolve(ctx, flagFrom)
					if err != nil {
						return fmt.Errorf("resolve --from: %w", err)
					}
				}

				svc := search.NewService(api, resolver, retryPolicy)

				result, err := svc.Search(ctx, search.SearchRequest{
					Peers:   peers,
					Query:   query,
					FromID:  fromPeer,
					Filter:  filter,
					After:   after,
					Before:  before,
					TopicID: flagTopic,
					Limit:   flagLimit,
					Cursor:  flagCursor,
				})
				if err != nil {
					return err
				}

				return writeSearchResult(cmd.OutOrStdout(), outputFormat(cmd), result)
			})
		},
	}

	cmd.Flags().StringSliceVarP(&flagChat, "chat", "c", nil, "chat to search (username, phone, ID; repeatable, comma-separated)")
	cmd.Flags().StringVar(&flagFolder, "folder", "", "search in chats of this folder (id or name); additive to --chat")
	cmd.Flags().StringVarP(&flagFrom, "from", "f", "", "filter by sender (username, phone, or ID)")
	cmd.Flags().StringVar(&flagFilter, "filter", "", "message type filter (photo, video, document, url, etc.)")
	cmd.Flags().StringVar(&flagAfter, "after", "", "only messages after date (YYYY-MM-DD or unix timestamp)")
	cmd.Flags().StringVar(&flagBefore, "before", "", "only messages before date (YYYY-MM-DD or unix timestamp)")
	cmd.Flags().IntVar(&flagTopic, "topic", 0, "forum topic ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "l", 50, "max messages to return (1-100)")
	cmd.Flags().StringVar(&flagCursor, "cursor", "", "pagination cursor from previous response")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer resolution cache")
	cmd.Flags().BoolVar(&flagIncludeComments, "include-comments", false, "also search the linked discussion group of each channel")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	return cmd
}

// expandChats splits each flag value by comma and trims spaces.
func expandChats(flags []string) []string {
	var out []string
	for _, f := range flags {
		for _, part := range strings.Split(f, ",") {
			s := strings.TrimSpace(part)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func validateLimit(limit int) error {
	if limit < 1 || limit > 100 {
		return fmt.Errorf("--limit must be between 1 and 100, got %d", limit)
	}
	return nil
}

// parseDate parses a date string: empty returns 0, YYYY-MM-DD or unix timestamp.
func parseDate(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	// Try unix timestamp first.
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return int(ts), nil
	}

	// Try YYYY-MM-DD.
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return 0, fmt.Errorf("expected YYYY-MM-DD or unix timestamp, got %q", s)
	}
	return int(t.Unix()), nil
}
