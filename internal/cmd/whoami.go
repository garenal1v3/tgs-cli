package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/searchtgcli/tgs/internal/profile"
	"github.com/searchtgcli/tgs/internal/telegram"
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

	client, err := telegram.New(profileName)
	if err != nil {
		return fmt.Errorf("open session: %w", err)
	}
	defer client.Close()

	info, err := client.LoadMeta()
	if err != nil {
		return fmt.Errorf("load user info: %w", err)
	}

	if flagOutput == "text" {
		fmt.Fprintf(cmd.OutOrStdout(), "Profile: %s\n", profileName)
		if info == nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Not logged in (no cached user info).")
			return nil
		}
		name := info.FirstName
		if info.LastName != "" {
			name += " " + info.LastName
		}
		fmt.Fprintf(cmd.OutOrStdout(), "User:    %s (id: %d)\n", name, info.ID)
		if info.Username != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Handle:  @%s\n", info.Username)
		}
		if info.Phone != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Phone:   %s\n", info.Phone)
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
