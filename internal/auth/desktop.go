package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	gosession "github.com/gotd/td/session"
	"github.com/gotd/td/session/tdesktop"
	"github.com/gotd/td/telegram"
)

// defaultTDataPaths returns the default Telegram Desktop tdata search paths for the
// current OS.
func defaultTDataPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Group Containers",
				"6N38VWS5BX.ru.keepcoder.Telegram", "appstore"),
			filepath.Join(home, "Library", "Application Support",
				"Telegram Desktop", "tdata"),
		}
	case "linux":
		return []string{
			filepath.Join(home, ".local", "share", "TelegramDesktop", "tdata"),
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return nil
		}
		return []string{
			filepath.Join(appData, "Telegram Desktop", "tdata"),
		}
	default:
		return nil
	}
}

// findTDataDir resolves the tdata directory: explicit override first, then
// auto-detection.
func findTDataDir(override string) (string, error) {
	if override != "" {
		if _, err := os.Stat(override); err != nil {
			return "", fmt.Errorf("tdata directory %q: %w", override, err)
		}
		return override, nil
	}

	for _, p := range defaultTDataPaths() {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("Telegram Desktop tdata not found; use --desktop-dir to specify the path")
}

// loginDesktop imports the first account from a Telegram Desktop tdata directory
// and stores it in storage so that the next client.Run() starts authenticated.
//
// The storage must be the same session.Storage that was passed to telegram.NewClient.
// This function must be called BEFORE client.Run(), or used with DesktopStorage to
// pre-populate the session.
func loginDesktop(ctx context.Context, _ *telegram.Client, storage gosession.Storage, desktopDir string, passcode string) error {
	tDataPath, err := findTDataDir(desktopDir)
	if err != nil {
		return err
	}

	var passcodeBytes []byte
	if passcode != "" {
		passcodeBytes = []byte(passcode)
	}

	accounts, err := tdesktop.Read(tDataPath, passcodeBytes)
	if err != nil {
		return fmt.Errorf("read Telegram Desktop session: %w", err)
	}
	if len(accounts) == 0 {
		return fmt.Errorf("no accounts found in Telegram Desktop tdata")
	}

	data, err := gosession.TDesktopSession(accounts[0])
	if err != nil {
		return fmt.Errorf("convert Telegram Desktop session: %w", err)
	}

	loader := gosession.Loader{Storage: storage}
	if err := loader.Save(ctx, data); err != nil {
		return fmt.Errorf("save imported session: %w", err)
	}

	return nil
}
