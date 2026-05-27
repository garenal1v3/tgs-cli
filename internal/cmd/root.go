package cmd

import (
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
		Use:   "tgs",
		Short: "Telegram search CLI",
		Long:  "Fast CLI client for searching Telegram via user account.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&flagProfile, "profile", "p", "default", "account profile name")
	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "json", "output format: json, text")
	root.PersistentFlags().BoolVar(&flagDebug, "debug", false, "enable debug logging")

	root.AddCommand(newVersionCmd())

	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
