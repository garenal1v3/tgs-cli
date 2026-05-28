package sources

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
	sourcessvc "github.com/searchtgcli/tgs/internal/sources"
	"github.com/searchtgcli/tgs/internal/telegram"
)

// newInspectCmd returns the "sources inspect" subcommand.
func newInspectCmd() *cobra.Command {
	var (
		flagNoStats bool
		flagNoCache bool
		flagMaxWait int
		flagProfile string
	)

	cmd := &cobra.Command{
		Use:   "inspect <ref>",
		Short: "Show detailed info about one source",
		Long:  "Show full info and metrics for a single source (channel, group, user). Works for unsubscribed channels via @username or phone.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]

			cwd, err := os.Getwd()
			if err != nil {
				return err
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

			meta, _ := client.LoadMeta()
			var selfID int64
			if meta != nil {
				selfID = meta.ID
			}

			return client.Run(cmd.Context(), func(ctx context.Context, api *tg.Client) error {
				retryPolicy := retry.DefaultPolicy()
				retryPolicy.MaxFloodWait = time.Duration(flagMaxWait) * time.Second

				var peerCache *resolve.PeerCache
				var statsCache *sourcessvc.SourceStatsCache
				if !flagNoCache {
					pc, err := resolve.NewPeerCache(telegram.CachePath(profileName))
					if err != nil {
						return fmt.Errorf("open peer cache: %w", err)
					}
					defer func() { _ = pc.Close() }()
					peerCache = pc

					sc, err := sourcessvc.NewSourceStatsCache(telegram.CachePath(profileName))
					if err != nil {
						return fmt.Errorf("open stats cache: %w", err)
					}
					defer func() { _ = sc.Close() }()
					statsCache = sc
				}

				resolver := resolve.NewResolver(api, peerCache)
				svc := sourcessvc.New(api, resolver, retryPolicy, statsCache, selfID)

				src, err := svc.Inspect(ctx, sourcessvc.InspectRequest{
					Ref:     ref,
					NoStats: flagNoStats,
				})
				if err != nil {
					return err
				}
				return writeSource(cmd.OutOrStdout(), outputFormat(cmd), src)
			})
		},
	}

	cmd.Flags().BoolVar(&flagNoStats, "no-stats", false, "skip expensive stats")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer and stats caches")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	return cmd
}
