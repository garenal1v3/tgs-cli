package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotd/td/telegram"
	tdauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func loginCode(ctx context.Context, client *telegram.Client, phone, code, password string) error {
	if phone == "" {
		var err error
		phone, err = prompt("Phone number: ")
		if err != nil {
			return fmt.Errorf("read phone: %w", err)
		}
	}

	var auth tdauth.UserAuthenticator
	if code != "" {
		auth = flagAuth{phone: phone, code: code, password: password}
	} else {
		auth = codeRequestAuth{phone: phone}
	}

	flow := tdauth.NewFlow(auth, tdauth.SendCodeOptions{})
	err := client.Auth().IfNecessary(ctx, flow)
	if err != nil && errors.Is(err, ErrCodeSent) {
		return ErrCodeSent
	}
	return err
}

// flagAuth implements UserAuthenticator using values from CLI flags (fully non-interactive).
type flagAuth struct {
	phone    string
	code     string
	password string
}

func (a flagAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a flagAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return a.code, nil
}

func (a flagAuth) Password(_ context.Context) (string, error) {
	if a.password == "" {
		return "", fmt.Errorf("2FA password required: use --password flag")
	}
	return a.password, nil
}

func (a flagAuth) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	return nil
}

func (a flagAuth) SignUp(_ context.Context) (tdauth.UserInfo, error) {
	return tdauth.UserInfo{}, fmt.Errorf("sign up not supported; use an existing Telegram account")
}

// codeRequestAuth sends the code and then returns ErrCodeSent from Code().
// Used for step 1 of the two-step non-interactive flow.
type codeRequestAuth struct {
	phone string
}

func (a codeRequestAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a codeRequestAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return "", ErrCodeSent
}

func (a codeRequestAuth) Password(_ context.Context) (string, error) {
	return "", ErrCodeSent
}

func (a codeRequestAuth) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	return nil
}

func (a codeRequestAuth) SignUp(_ context.Context) (tdauth.UserInfo, error) {
	return tdauth.UserInfo{}, fmt.Errorf("sign up not supported; use an existing Telegram account")
}
