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

// newListCmd returns the "sources list" subcommand.
func newListCmd() *cobra.Command {
	var (
		flagType    []string
		flagStats   bool
		flagLimit   int
		flagCursor  string
		flagArchive bool
		flagMaxWait int
		flagNoCache bool
		flagProfile string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sources available in the account",
		Long:  "Enumerate channels, supergroups, groups, users and bots from the active profile's dialogs.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			types, err := parseTypeFilter(flagType)
			if err != nil {
				return err
			}
			if flagLimit < 0 || flagLimit > 500 {
				return fmt.Errorf("--limit must be 0..500, got %d", flagLimit)
			}

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

					sc, err := sourcessvc.NewSourceStatsCache(telegram.StatsCachePath(profileName))
					if err != nil {
						return fmt.Errorf("open stats cache: %w", err)
					}
					defer func() { _ = sc.Close() }()
					statsCache = sc
				}

				resolver := resolve.NewResolver(api, peerCache)
				svc := sourcessvc.New(api, resolver, retryPolicy, statsCache, selfID)

				result, err := svc.List(ctx, sourcessvc.ListRequest{
					Types:     types,
					WithStats: flagStats,
					Limit:     flagLimit,
					Cursor:    flagCursor,
					Archived:  flagArchive,
				})
				if err != nil {
					return err
				}
				return writeSourceList(cmd.OutOrStdout(), outputFormat(cmd), result)
			})
		},
	}

	cmd.Flags().StringSliceVar(&flagType, "type", nil, "filter by type: channel,supergroup,group,user,bot (repeatable, comma-separated)")
	cmd.Flags().BoolVar(&flagStats, "with-stats", false, "fetch expensive metrics (total/24h/first) and full info per source")
	cmd.Flags().IntVarP(&flagLimit, "limit", "l", 0, "max records (1-500, 0=all)")
	cmd.Flags().StringVar(&flagCursor, "cursor", "", "pagination cursor from previous response")
	cmd.Flags().BoolVar(&flagArchive, "archived", false, "include archived dialogs")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer and stats caches")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	return cmd
}
