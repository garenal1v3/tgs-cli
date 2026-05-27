package search

import "github.com/spf13/cobra"

// NewSearchCmd creates the parent "search" command with all subcommands.
func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search messages in Telegram",
		Long:  "Search messages across chats, channels, and groups using Telegram MTProto API.",
	}
	cmd.AddCommand(
		newMessagesCmd(),
		newGlobalCmd(),
		newCountersCmd(),
		newCalendarCmd(),
	)
	return cmd
}
