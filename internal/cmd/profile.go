package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/searchtgcli/tgs/internal/profile"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage account profiles",
	}

	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileSwitchCmd())
	cmd.AddCommand(newProfileDeleteCmd())

	return cmd
}

// profile list

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		RunE:  runProfileList,
	}
}

func runProfileList(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	active := profile.Resolve(flagProfile, cwd)

	profiles, err := profile.List()
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}

	if flagOutput == "text" {
		if len(profiles) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No profiles found.")
			return nil
		}
		for _, p := range profiles {
			marker := "  "
			if p.Name == active {
				marker = "* "
			}
			line := marker + p.Name
			if p.Username != "" {
				line += " (@" + p.Username + ")"
			} else if p.Phone != "" {
				line += " (" + p.Phone + ")"
			}
			fmt.Fprintln(cmd.OutOrStdout(), line)
		}
		return nil
	}

	type profileJSON struct {
		Name     string `json:"name"`
		Phone    string `json:"phone,omitempty"`
		Username string `json:"username,omitempty"`
		UserID   string `json:"user_id,omitempty"`
		Active   bool   `json:"active"`
	}
	result := make([]profileJSON, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, profileJSON{
			Name:     p.Name,
			Phone:    p.Phone,
			Username: p.Username,
			UserID:   p.UserID,
			Active:   p.Name == active,
		})
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
}

// profile switch

func newProfileSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Set the active profile for the current directory",
		Args:  cobra.ExactArgs(1),
		RunE:  runProfileSwitch,
	}
}

func runProfileSwitch(cmd *cobra.Command, args []string) error {
	name := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if err := profile.Switch(name, cwd); err != nil {
		return fmt.Errorf("switch profile: %w", err)
	}

	if flagOutput == "text" {
		fmt.Fprintf(cmd.OutOrStdout(), "Switched to profile %q (wrote .tgs.yaml in %s)\n", name, cwd)
		return nil
	}

	return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
		"profile": name,
		"status":  "switched",
		"dir":     cwd,
	})
}

// profile delete

func newProfileDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a profile and its stored session",
		Args:  cobra.ExactArgs(1),
		RunE:  runProfileDelete,
	}
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	active := profile.Resolve(flagProfile, cwd)
	if name == active {
		return fmt.Errorf("cannot delete the currently active profile %q; switch to another profile first", name)
	}

	if err := profile.Delete(name); err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	if flagOutput == "text" {
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted profile %q\n", name)
		return nil
	}

	return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
		"profile": name,
		"status":  "deleted",
	})
}
