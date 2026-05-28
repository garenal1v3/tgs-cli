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

func newCountersCmd() *cobra.Command {
	var (
		flagChat    string
		flagTopic   int
		flagFilters []string
		flagMaxWait int
		flagNoCache bool
		flagProfile string
	)

	cmd := &cobra.Command{
		Use:   "counters",
		Short: "Get message counts by type",
		Long:  "Get message counts grouped by filter type (photo, video, document, etc.) for a specific chat.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Parse filter names if provided.
			var filters []tg.MessagesFilterClass
			for _, name := range flagFilters {
				f, err := search.ParseFilter(name)
				if err != nil {
					return err
				}
				filters = append(filters, f)
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

				result, err := svc.GetCounters(ctx, search.CountersRequest{
					Peer:    peer,
					TopicID: flagTopic,
					Filters: filters,
				})
				if err != nil {
					return err
				}

				return writeCountersResult(cmd.OutOrStdout(), outputFormat(cmd), result)
			})
		},
	}

	cmd.Flags().StringVarP(&flagChat, "chat", "c", "", "chat to query (username, phone, or ID)")
	cmd.Flags().IntVar(&flagTopic, "topic", 0, "forum topic ID")
	cmd.Flags().StringSliceVar(&flagFilters, "filters", nil, "filter types to count (comma-separated; default: all)")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer resolution cache")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	_ = cmd.MarkFlagRequired("chat")

	return cmd
}
