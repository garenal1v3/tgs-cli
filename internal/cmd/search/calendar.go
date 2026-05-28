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

func newCalendarCmd() *cobra.Command {
	var (
		flagChat    string
		flagFolder  string
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

			if flagChat == "" && flagFolder == "" {
				return fmt.Errorf("at least one of --chat or --folder is required")
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
				svc := search.NewService(api, resolver, retryPolicy)

				var peers []tg.InputPeerClass
				if flagChat != "" {
					p, err := resolver.Resolve(ctx, flagChat)
					if err != nil {
						return fmt.Errorf("resolve --chat: %w", err)
					}
					peers = append(peers, p)
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
					return fmt.Errorf("no chats to query")
				}

				if len(peers) == 1 && flagFolder == "" {
					result, err := svc.GetCalendar(ctx, search.CalendarRequest{Peer: peers[0], Filter: filter})
					if err != nil {
						return err
					}
					return writeCalendarResult(cmd.OutOrStdout(), outputFormat(cmd), result)
				}

				multi := &search.MultiCalendarResult{}
				for _, peer := range peers {
					res, err := svc.GetCalendar(ctx, search.CalendarRequest{Peer: peer, Filter: filter})
					if err != nil {
						return fmt.Errorf("calendar for peer %d: %w", peerID(peer), err)
					}
					multi.Chats = append(multi.Chats, search.PerChatCalendar{
						Chat:    search.ChatRef{ID: peerID(peer)},
						Periods: res.Periods,
						Total:   res.Total,
					})
				}
				return writeMultiCalendarResult(cmd.OutOrStdout(), outputFormat(cmd), multi)
			})
		},
	}

	cmd.Flags().StringVarP(&flagChat, "chat", "c", "", "chat to query (username, phone, or ID)")
	cmd.Flags().StringVar(&flagFolder, "folder", "", "fan-out calendar across chats of this folder (id or name)")
	cmd.Flags().StringVar(&flagFilter, "filter", "", "message type filter (photo, video, document, url, etc.)")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer resolution cache")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	_ = cmd.MarkFlagRequired("filter")

	return cmd
}
