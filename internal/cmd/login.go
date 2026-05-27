package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/searchtgcli/tgs/internal/auth"
	"github.com/searchtgcli/tgs/internal/profile"
	"github.com/searchtgcli/tgs/internal/telegram"
	tg "github.com/gotd/td/tg"
)

var (
	flagLoginType   string
	flagDesktopDir  string
	flagPasscode    string
	flagPhone       string
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Telegram",
		Long: `Authenticate with Telegram using one of three methods:

  desktop  Import session from Telegram Desktop (default)
  code     Authenticate via phone number and SMS/app code
  qr       Authenticate by scanning a QR code`,
		RunE: runLogin,
	}

	cmd.Flags().StringVarP(&flagLoginType, "type", "T", "desktop", "auth method: desktop, code, qr")
	cmd.Flags().StringVarP(&flagDesktopDir, "desktop-dir", "d", "", "custom Telegram Desktop tdata path")
	cmd.Flags().StringVar(&flagPasscode, "passcode", "", "Telegram Desktop local passcode")
	cmd.Flags().StringVar(&flagPhone, "phone", "", "phone number for code auth (e.g. +12345678900)")

	return cmd
}

func runLogin(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	profileName := profile.Resolve(flagProfile, cwd)

	method, err := auth.ParseMethod(flagLoginType)
	if err != nil {
		return err
	}

	client, err := telegram.New(profileName)
	if err != nil {
		return fmt.Errorf("create telegram client: %w", err)
	}
	defer client.Close()

	opts := auth.Options{
		Method:     method,
		Phone:      flagPhone,
		DesktopDir: flagDesktopDir,
		Passcode:   flagPasscode,
	}

	if method == auth.MethodDesktop {
		// Desktop: import session BEFORE Run(), then Run() to verify and fetch user info.
		opts.Storage = client.SessionStorage()
		if err := auth.Login(context.Background(), client.RawClient(), opts); err != nil {
			return fmt.Errorf("import desktop session: %w", err)
		}

		var userInfo *telegram.UserInfo
		if err := client.Run(context.Background(), func(ctx context.Context, api *tg.Client) error {
			info, fetchErr := fetchUserInfo(ctx, api)
			if fetchErr != nil {
				return fetchErr
			}
			userInfo = info
			return client.StoreMeta(info)
		}); err != nil {
			return fmt.Errorf("connect and verify session: %w", err)
		}

		return printLoginResult(cmd, profileName, userInfo)
	}

	// Code/QR: call auth.Login inside Run().
	var userInfo *telegram.UserInfo
	if err := client.Run(context.Background(), func(ctx context.Context, api *tg.Client) error {
		if err := auth.Login(ctx, client.RawClient(), opts); err != nil {
			return err
		}
		info, fetchErr := fetchUserInfo(ctx, api)
		if fetchErr != nil {
			return fetchErr
		}
		userInfo = info
		return client.StoreMeta(info)
	}); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	return printLoginResult(cmd, profileName, userInfo)
}

// fetchUserInfo retrieves basic user info from the current session.
func fetchUserInfo(ctx context.Context, api *tg.Client) (*telegram.UserInfo, error) {
	users, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("no user returned by UsersGetUsers")
	}
	user, ok := users[0].(*tg.User)
	if !ok {
		return nil, fmt.Errorf("unexpected user type: %T", users[0])
	}
	return &telegram.UserInfo{
		ID:        user.ID,
		Phone:     user.Phone,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

func printLoginResult(cmd *cobra.Command, profileName string, info *telegram.UserInfo) error {
	if flagOutput == "text" {
		if info != nil {
			name := info.FirstName
			if info.LastName != "" {
				name += " " + info.LastName
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s (id: %d, profile: %s)\n", name, info.ID, profileName)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Logged in (profile: %s)\n", profileName)
		}
		return nil
	}

	out := map[string]interface{}{
		"profile": profileName,
		"status":  "logged_in",
	}
	if info != nil {
		out["user"] = info
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
}
