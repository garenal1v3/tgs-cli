package auth

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	tdauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// loginCode performs authentication via phone number + SMS/app code with optional
// 2FA password. Must be called from inside client.Run().
func loginCode(ctx context.Context, client *telegram.Client, phone string) error {
	if phone == "" {
		var err error
		phone, err = prompt("Phone number: ")
		if err != nil {
			return fmt.Errorf("read phone: %w", err)
		}
	}

	flow := tdauth.NewFlow(
		terminalAuth{phone: phone},
		tdauth.SendCodeOptions{},
	)

	return client.Auth().IfNecessary(ctx, flow)
}

// terminalAuth implements tdauth.UserAuthenticator by prompting the terminal.
type terminalAuth struct {
	phone string
}

func (a terminalAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a terminalAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return prompt("Code: ")
}

func (a terminalAuth) Password(_ context.Context) (string, error) {
	return promptPassword("2FA Password: ")
}

func (a terminalAuth) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	return nil
}

func (a terminalAuth) SignUp(_ context.Context) (tdauth.UserInfo, error) {
	return tdauth.UserInfo{}, fmt.Errorf("sign up not supported; use an existing Telegram account")
}
