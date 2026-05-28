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

func newFoldersCmd() *cobra.Command {
	var (
		flagArchived bool
		flagMaxWait  int
		flagNoCache  bool
		flagProfile  string
	)

	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List Telegram folders and their contents",
		Long:  "Enumerate the user's dialog filters (folders) and resolve each one's contents the way Telegram's UI does (applying pinned/include/exclude peers and auto-flags).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
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

			return client.Run(cmd.Context(), func(ctx context.Context, api *tg.Client) error {
				retryPolicy := retry.DefaultPolicy()
				retryPolicy.MaxFloodWait = time.Duration(flagMaxWait) * time.Second

				var peerCache *resolve.PeerCache
				if !flagNoCache {
					pc, err := resolve.NewPeerCache(telegram.CachePath(profileName))
					if err != nil {
						return fmt.Errorf("open peer cache: %w", err)
					}
					defer func() { _ = pc.Close() }()
					peerCache = pc
				}

				resolver := resolve.NewResolver(api, peerCache)
				svc := sourcessvc.New(api, resolver, retryPolicy, nil, selfID)

				result, err := svc.Folders(ctx, sourcessvc.FoldersRequest{Archived: flagArchived})
				if err != nil {
					return err
				}
				return writeFoldersResult(cmd.OutOrStdout(), outputFormat(cmd), result)
			})
		},
	}

	cmd.Flags().BoolVar(&flagArchived, "archived", false, "include archived dialogs in folder contents")
	cmd.Flags().IntVar(&flagMaxWait, "max-wait", 60, "max seconds to wait on FLOOD_WAIT")
	cmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "disable peer cache")
	cmd.Flags().StringVarP(&flagProfile, "profile", "p", "", "account profile name")

	return cmd
}
