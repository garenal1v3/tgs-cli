package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagOutput == "text" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "tgs %s\ncommit: %s\nbuilt:  %s\n", Version, Commit, Date)
				return nil
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
				"version": Version,
				"commit":  Commit,
				"date":    Date,
			})
		},
	}
}
