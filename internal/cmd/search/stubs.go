package search

import "github.com/spf13/cobra"

// Stub commands for subcommands not yet implemented (Task 10).

func newGlobalCmd() *cobra.Command {
	return &cobra.Command{Use: "global", Hidden: true}
}

func newCountersCmd() *cobra.Command {
	return &cobra.Command{Use: "counters", Hidden: true}
}

func newCalendarCmd() *cobra.Command {
	return &cobra.Command{Use: "calendar", Hidden: true}
}
