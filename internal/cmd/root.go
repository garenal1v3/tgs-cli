package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagProfile string
	flagOutput  string
	flagDebug   bool
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
	root.PersistentFlags().BoolVar(&flagDebug, "debug", false, "enable debug logging")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newLoginCmd())
	root.AddCommand(newLogoutCmd())
	root.AddCommand(newProfileCmd())
	root.AddCommand(newWhoamiCmd())

	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
