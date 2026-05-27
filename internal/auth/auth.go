package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	gosession "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"golang.org/x/term"
)

// Method represents an authentication method.
type Method int

const (
	MethodDesktop Method = iota
	MethodCode
	MethodQR
)

// ParseMethod parses an auth method string.
func ParseMethod(s string) (Method, error) {
	switch s {
	case "desktop", "":
		return MethodDesktop, nil
	case "code":
		return MethodCode, nil
	case "qr":
		return MethodQR, nil
	default:
		return 0, fmt.Errorf("unknown auth method: %q (use desktop, code, or qr)", s)
	}
}

// Options holds parameters for the Login function.
type Options struct {
	Method     Method
	Phone      string
	DesktopDir string
	Passcode   string

	// Storage is required for MethodDesktop: it must be the same session.Storage
	// that was passed to telegram.NewClient so the imported session is persisted
	// in the client's session store.
	Storage gosession.Storage

	// LoggedIn is an optional channel for MethodQR. When non-nil it receives a
	// signal when the Telegram server sends UpdateLoginToken (i.e. the QR was
	// scanned). Callers should create it via qrlogin.OnLoginToken(dispatcher)
	// and pass the same dispatcher as telegram.Options.UpdateHandler when
	// building the client. When nil, a polling fallback is used instead.
	LoggedIn qrlogin.LoggedIn
}

// Login authenticates the client using the specified method.
//
// For MethodDesktop, Login writes the imported session to opts.Storage and
// returns; the caller must call client.Run() to connect with the new session.
// For MethodCode and MethodQR, Login must be called from inside client.Run().
func Login(ctx context.Context, client *telegram.Client, opts Options) error {
	switch opts.Method {
	case MethodDesktop:
		return loginDesktop(ctx, client, opts.Storage, opts.DesktopDir, opts.Passcode)
	case MethodCode:
		return loginCode(ctx, client, opts.Phone)
	case MethodQR:
		return loginQR(ctx, client, opts.LoggedIn)
	default:
		return fmt.Errorf("unsupported auth method: %d", opts.Method)
	}
}

// prompt writes label to stderr and reads one line from stdin.
func prompt(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("no input")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

// promptPassword writes label to stderr and reads a password without echo.
func promptPassword(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(pw), nil
}
