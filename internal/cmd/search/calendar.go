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
	"github.com/searchtgcli/tgs/internal/telegram"
)

func newCalendarCmd() *cobra.Command {
	var (
		flagChat    string
		flagFilter  string
		flagMaxWait int
		flagNoCache bool
		flagProfile string
	)

	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Get search results grouped by date",
		Long:  "Get message search results grouped by date for a specific chat and filter type.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			filter, err := search.ParseFilter(flagFilter)
			if err != nil {
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

				peer, err := resolver.Resolve(ctx, flagChat)
				if err != nil {
					return fmt.Errorf("resolve --chat: %w", err)
				}

				svc := search.NewService(api, resolver, retryPolicy)

				result, err := svc.GetCalendar(ctx, search.CalendarRequest{
					Peer:   peer,
					Filter: filter,
				})
				if err != nil {
					return err
				}

				return writeCalendarResult(cmd.OutOrStdout(), outputFormat(cmd), result)
			})
		},
	}

	cmd.Flags().StringVarP(&flagChat, "chat", "c", "", "chat to query (username, phone, or ID)")
	cmd.Flags().StringVar(&flagFilter, "filter", "", "message type filter (photo, video, document, url, etc.)")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer resolution cache")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	_ = cmd.MarkFlagRequired("chat")
	_ = cmd.MarkFlagRequired("filter")

	return cmd
}
