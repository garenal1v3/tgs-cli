package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/garenal1v3/tgs-cli/internal/profile"
	"github.com/garenal1v3/tgs-cli/internal/telegram"
)

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the current profile and authenticated user",
		Long:  "Shows the active profile name and the cached user info (no Telegram connection required).",
		RunE:  runWhoami,
	}
}

func runWhoami(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	profileName := profile.Resolve(flagProfile, cwd)

	client, err := telegram.Open(profileName)
	if err != nil {
		return fmt.Errorf("open session: %w", err)
	}

	var info *telegram.UserInfo
	if client != nil {
		defer func() { _ = client.Close() }()
		info, err = client.LoadMeta()
		if err != nil {
			return fmt.Errorf("load user info: %w", err)
		}
	}

	if flagOutput == "text" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Profile: %s\n", profileName)
		if info == nil {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Not logged in (no cached user info).")
			return nil
		}
		name := info.FirstName
		if info.LastName != "" {
			name += " " + info.LastName
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User:    %s (id: %d)\n", name, info.ID)
		if info.Username != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Handle:  @%s\n", info.Username)
		}
		if info.Phone != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Phone:   %s\n", info.Phone)
		}
		return nil
	}

	out := map[string]interface{}{
		"profile": profileName,
	}
	if info != nil {
		out["user"] = info
	} else {
		out["user"] = nil
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
}
