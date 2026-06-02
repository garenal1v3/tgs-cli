package search

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/garenal1v3/tgs-cli/internal/profile"
	"github.com/garenal1v3/tgs-cli/internal/resolve"
	"github.com/garenal1v3/tgs-cli/internal/retry"
	"github.com/garenal1v3/tgs-cli/internal/search"
	"github.com/garenal1v3/tgs-cli/internal/sources"
	"github.com/garenal1v3/tgs-cli/internal/telegram"
)

func newCountersCmd() *cobra.Command {
	var (
		flagChat    string
		flagFolder  string
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

				var refs []search.ChatRef
				var peers []tg.InputPeerClass
				if flagChat != "" {
					p, err := resolver.Resolve(ctx, flagChat)
					if err != nil {
						return fmt.Errorf("resolve --chat: %w", err)
					}
					peers = append(peers, p)
					refs = append(refs, chatRefFromPeer(p))
				}
				if flagFolder != "" {
					meta, _ := client.LoadMeta()
					var selfID int64
					if meta != nil {
						selfID = meta.ID
					}
					src := sources.New(api, resolver, retryPolicy, nil, selfID)
					resolved, err := src.ResolveFolder(ctx, flagFolder)
					if err != nil {
						return fmt.Errorf("resolve --folder: %w", err)
					}
					for i, c := range resolved.Folder.Chats {
						peers = append(peers, resolved.Peers[i])
						refs = append(refs, chatRefFromSource(c))
					}
				}
				if len(peers) == 0 {
					return fmt.Errorf("no chats to query: --folder %q resolved to no chats and no --chat was given", flagFolder)
				}

				if len(peers) == 1 && flagFolder == "" {
					// single-chat path: keep legacy output format
					result, err := svc.GetCounters(ctx, search.CountersRequest{
						Peer:    peers[0],
						TopicID: flagTopic,
						Filters: filters,
					})
					if err != nil {
						return err
					}
					return writeCountersResult(cmd.OutOrStdout(), outputFormat(cmd), result)
				}

				// Multi-chat: per-chat breakdown + aggregated totals (kept in
				// first-seen filter order so the output is stable, not map-random).
				multi := &search.MultiCountersResult{}
				totals := make(map[string]int)
				var totalOrder []string
				for i, peer := range peers {
					res, err := svc.GetCounters(ctx, search.CountersRequest{
						Peer:    peer,
						TopicID: flagTopic,
						Filters: filters,
					})
					if err != nil {
						return fmt.Errorf("counters for peer %d: %w", peerID(peer), err)
					}
					multi.Chats = append(multi.Chats, search.PerChatCounters{
						Chat:     refs[i],
						Counters: res.Counters,
					})
					for _, c := range res.Counters {
						if _, seen := totals[c.Filter]; !seen {
							totalOrder = append(totalOrder, c.Filter)
						}
						totals[c.Filter] += c.Count
					}
				}
				for _, name := range totalOrder {
					multi.Totals = append(multi.Totals, search.CounterEntry{Filter: name, Count: totals[name]})
				}
				return writeMultiCountersResult(cmd.OutOrStdout(), outputFormat(cmd), multi)
			})
		},
	}

	cmd.Flags().StringVarP(&flagChat, "chat", "c", "", "chat to query (username, phone, or ID)")
	cmd.Flags().StringVar(&flagFolder, "folder", "", "fan-out counters across chats of this folder (id or name)")
	cmd.Flags().IntVar(&flagTopic, "topic", 0, "forum topic ID")
	cmd.Flags().StringSliceVar(&flagFilters, "filters", nil, "filter types to count (comma-separated; default: all)")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer resolution cache")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	return cmd
}

func peerID(p tg.InputPeerClass) int64 {
	switch v := p.(type) {
	case *tg.InputPeerUser:
		return v.UserID
	case *tg.InputPeerChat:
		return v.ChatID
	case *tg.InputPeerChannel:
		return v.ChannelID
	}
	return 0
}
