package search

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/searchtgcli/tgs/internal/profile"
	"github.com/searchtgcli/tgs/internal/retry"
	"github.com/searchtgcli/tgs/internal/search"
	"github.com/searchtgcli/tgs/internal/telegram"
)

func newGlobalCmd() *cobra.Command {
	var (
		flagChannelsOnly bool
		flagGroupsOnly   bool
		flagUsersOnly    bool
		flagFolder       int
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

			client, err := telegram.Open(profileName)
			if err != nil {
				return fmt.Errorf("open session: %w", err)
			}
			if client == nil {
				return fmt.Errorf("run tgs login first")
			}
			defer func() { _ = client.Close() }()

			ctx := cmd.Context()
			return client.Run(ctx, func(ctx context.Context, api *tg.Client) error {
				retryPolicy := retry.DefaultPolicy()
				retryPolicy.MaxFloodWait = time.Duration(flagMaxWait) * time.Second

				svc := search.NewService(api, nil, retryPolicy)

				result, err := svc.SearchGlobal(ctx, search.GlobalSearchRequest{
					Query:        query,
					Filter:       filter,
					After:        after,
					Before:       before,
					FolderID:     flagFolder,
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
	cmd.Flags().IntVar(&flagFolder, "folder", 0, "search only in folder with this ID")
	cmd.Flags().StringVar(&flagFilter, "filter", "", "message type filter (photo, video, document, url, etc.)")
	cmd.Flags().StringVar(&flagAfter, "after", "", "only messages after date (YYYY-MM-DD or unix timestamp)")
	cmd.Flags().StringVar(&flagBefore, "before", "", "only messages before date (YYYY-MM-DD or unix timestamp)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "l", 50, "max messages to return (1-100)")
	cmd.Flags().StringVar(&flagCursor, "cursor", "", "pagination cursor from previous response")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	return cmd
}
