package search

import (
	"context"
	"fmt"
	"os"
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

func newGlobalCmd() *cobra.Command {
	var (
		flagChannelsOnly bool
		flagGroupsOnly   bool
		flagUsersOnly    bool
		flagArchived     bool
		flagFolder       string
		flagFilter       string
		flagAfter        string
		flagBefore       string
		flagLimit        int
		flagCursor       string
		flagMaxWait      int
		flagProfile      string
	)

	cmd := &cobra.Command{
		Use:   "global [query]",
		Short: "Search messages across all chats",
		Long:  "Search messages globally across all chats, channels, and groups.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			typeFlags := 0
			if flagChannelsOnly {
				typeFlags++
			}
			if flagGroupsOnly {
				typeFlags++
			}
			if flagUsersOnly {
				typeFlags++
			}
			if typeFlags > 1 {
				return fmt.Errorf("--channels-only, --groups-only, and --users-only are mutually exclusive")
			}
			if flagFolder != "" && (flagChannelsOnly || flagGroupsOnly || flagUsersOnly) {
				return fmt.Errorf("--folder cannot be combined with --channels-only/--groups-only/--users-only")
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

			meta, _ := client.LoadMeta()
			var selfID int64
			if meta != nil {
				selfID = meta.ID
			}

			ctx := cmd.Context()
			return client.Run(ctx, func(ctx context.Context, api *tg.Client) error {
				retryPolicy := retry.DefaultPolicy()
				retryPolicy.MaxFloodWait = time.Duration(flagMaxWait) * time.Second

				svc := search.NewService(api, nil, retryPolicy)

				if flagFolder != "" {
					// Folder-scoped search: resolve folder peers and route through the
					// multi-peer search path (messages.searchGlobal can't accept a peer
					// list, so we fan out via sources.Service.ResolveFolder + search.Service.Search).
					var peerCache *resolve.PeerCache
					if pc, err := resolve.NewPeerCache(telegram.CachePath(profileName)); err == nil {
						peerCache = pc
						defer func() { _ = pc.Close() }()
					}
					resolver := resolve.NewResolver(api, peerCache)
					src := sources.New(api, resolver, retryPolicy, nil, selfID)
					// --archived is ignored when --folder is set: a folder view
					// always spans the archive (ResolveFolder walks it).
					resolved, err := src.ResolveFolder(ctx, flagFolder)
					if err != nil {
						return fmt.Errorf("resolve --folder: %w", err)
					}
					if len(resolved.Peers) == 0 {
						return fmt.Errorf("no chats to search: --folder %q resolved to no chats", flagFolder)
					}
					// Replace svc with one that has the resolver (search.NewService accepts nil for
					// resolver in the global path, but Search may need it for peer-id round-trip).
					svc = search.NewService(api, resolver, retryPolicy)
					result, err := svc.Search(ctx, search.SearchRequest{
						Peers:  resolved.Peers,
						Query:  query,
						Filter: filter,
						After:  after,
						Before: before,
						Limit:  flagLimit,
						Cursor: flagCursor,
					})
					if err != nil {
						return err
					}
					return writeSearchResult(cmd.OutOrStdout(), outputFormat(cmd), result)
				}

				result, err := svc.SearchGlobal(ctx, search.GlobalSearchRequest{
					Query:        query,
					Filter:       filter,
					After:        after,
					Before:       before,
					Archived:     flagArchived,
					ChannelsOnly: flagChannelsOnly,
					GroupsOnly:   flagGroupsOnly,
					UsersOnly:    flagUsersOnly,
					Limit:        flagLimit,
					Cursor:       flagCursor,
				})
				if err != nil {
					return err
				}

				return writeSearchResult(cmd.OutOrStdout(), outputFormat(cmd), result)
			})
		},
	}

	cmd.Flags().BoolVar(&flagChannelsOnly, "channels-only", false, "search only in channels")
	cmd.Flags().BoolVar(&flagGroupsOnly, "groups-only", false, "search only in groups")
	cmd.Flags().BoolVar(&flagUsersOnly, "users-only", false, "search only in private chats")
	cmd.Flags().BoolVar(&flagArchived, "archived", false, "search in the archive folder instead of main")
	cmd.Flags().StringVar(&flagFolder, "folder", "", "search only in chats of this folder (id or name)")
	cmd.Flags().StringVar(&flagFilter, "filter", "", "message type filter (photo, video, document, url, etc.)")
	cmd.Flags().StringVar(&flagAfter, "after", "", "only messages after date (YYYY-MM-DD or unix timestamp)")
	cmd.Flags().StringVar(&flagBefore, "before", "", "only messages before date (YYYY-MM-DD or unix timestamp)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "l", 50, "max messages to return (1-100)")
	cmd.Flags().StringVar(&flagCursor, "cursor", "", "pagination cursor from previous response")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	return cmd
}
