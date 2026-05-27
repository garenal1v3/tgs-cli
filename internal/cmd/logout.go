package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	tg "github.com/gotd/td/tg"

	"github.com/searchtgcli/tgs/internal/profile"
	"github.com/searchtgcli/tgs/internal/telegram"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out from Telegram and clear the session",
		RunE:  runLogout,
	}
}

func runLogout(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	profileName := profile.Resolve(flagProfile, cwd)

	client, err := telegram.Open(profileName)
	if err != nil {
		return fmt.Errorf("open session: %w", err)
	}
	if client == nil {
		return fmt.Errorf("profile %q is not logged in", profileName)
	}

	if err := client.Run(context.Background(), func(ctx context.Context, api *tg.Client) error {
		_, err := api.AuthLogOut(ctx)
		return err
	}); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	_ = client.Close()
	if err := profile.Delete(profileName); err != nil {
		return fmt.Errorf("clear local session: %w", err)
	}

	if flagOutput == "text" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Logged out (profile: %s)\n", profileName)
		return nil
	}

	return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
		"profile": profileName,
		"status":  "logged_out",
	})
}
