package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	searchcmd "github.com/garenal1v3/tgs-cli/internal/cmd/search"
	sourcescmd "github.com/garenal1v3/tgs-cli/internal/cmd/sources"
)

var (
	flagProfile string
	flagOutput  string
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "tgs",
		Short:         "Telegram search CLI",
		Long:          "Fast CLI client for searching Telegram via user account.",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if flagOutput != "json" && flagOutput != "text" {
				return fmt.Errorf("unknown output format %q (use json or text)", flagOutput)
			}
			return nil
		},
	}

	root.PersistentFlags().StringVarP(&flagProfile, "profile", "p", "", "account profile name (default: TGS_PROFILE > .tgs.yaml > \"default\")")
	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "json", "output format: json, text")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newLoginCmd())
	root.AddCommand(newLogoutCmd())
	root.AddCommand(newProfileCmd())
	root.AddCommand(newWhoamiCmd())
	root.AddCommand(searchcmd.NewSearchCmd())
	root.AddCommand(sourcescmd.NewSourcesCmd())

	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
