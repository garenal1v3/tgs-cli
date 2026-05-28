package sources

import "github.com/spf13/cobra"

// NewSourcesCmd creates the "sources" command with list/inspect subcommands.
func NewSourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Inspect chats, channels, groups and users in the account",
		Long:  "Enumerate Telegram dialogs available in the active profile and inspect specific peers.",
	}
	cmd.AddCommand(newListCmd(), newInspectCmd())
	return cmd
}

func newListCmd() *cobra.Command    { return &cobra.Command{Use: "list", Hidden: true} }
func newInspectCmd() *cobra.Command { return &cobra.Command{Use: "inspect", Hidden: true} }
