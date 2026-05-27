package auth

import (
	"context"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
)

// loginQR performs authentication via QR code scan. Must be called from inside
// client.Run().
//
// loggedIn is an optional channel that fires when the Telegram server sends an
// UpdateLoginToken (requires the client to have been created with an
// UpdateHandler that wires up the channel via qrlogin.OnLoginToken). When nil,
// the function falls back to a polling loop that re-exports the token on a
// timer until the QR is scanned or the context is cancelled.
func loginQR(ctx context.Context, client *telegram.Client, loggedIn qrlogin.LoggedIn) error {
	// Check whether already authorized — skip QR flow if so.
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("auth status: %w", err)
	}
	if status.Authorized {
		return nil
	}

	qr := client.QR()

	if loggedIn != nil {
		// Push-based path: server sends UpdateLoginToken when QR is scanned.
		_, err = qr.Auth(ctx, loggedIn, func(ctx context.Context, token qrlogin.Token) error {
			fmt.Fprintln(os.Stderr, "\nScan the QR code with your Telegram app:")
			printTokenQR(token)
			return nil
		})
		return err
	}

	// Polling fallback: repeatedly export the QR token and show it. After each
	// export wait up to its expiry, polling periodically via a fresh Export.
	// An empty token from Export means AuthLoginTokenSuccess — scan detected.
	const pollInterval = 3 * time.Second

	for {
		token, exportErr := qr.Export(ctx)
		if exportErr != nil {
			return fmt.Errorf("export QR token: %w", exportErr)
		}

		if token.Empty() {
			// QR was already scanned — finalise auth.
			_, err = qr.Import(ctx)
			return err
		}

		fmt.Fprintln(os.Stderr, "\nScan the QR code with your Telegram app:")
		printTokenQR(token)
		fmt.Fprintln(os.Stderr, "(Waiting for scan... The code refreshes automatically.)")

		// Wait until token expires (or poll faster), checking for a scanned signal
		// on every tick.
		deadline := token.Expires()
		if deadline.IsZero() || deadline.Before(time.Now()) {
			deadline = time.Now().Add(30 * time.Second)
		}

		for time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(pollInterval):
			}

			// Re-export: if we get an empty token the QR was scanned.
			t, exportErr := qr.Export(ctx)
			if exportErr != nil {
				return fmt.Errorf("export QR token: %w", exportErr)
			}
			if t.Empty() {
				_, err = qr.Import(ctx)
				return err
			}
			// Not yet scanned; the new token may have a different expiry.
			token = t
			deadline = token.Expires()
			if deadline.IsZero() || deadline.Before(time.Now()) {
				deadline = time.Now().Add(30 * time.Second)
			}
		}
		// Token expired — outer loop re-exports a fresh one.
	}
}

// printTokenQR renders the QR code for token to stderr. Falls back to URL on
// render failure.
func printTokenQR(token qrlogin.Token) {
	img, err := token.Image(0) // 0 == qr.L (smallest / least redundant)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QR render failed (%v); open this URL manually:\n%s\n", err, token.URL())
		return
	}
	printQRImage(img)
}

// printQRImage renders an image.Image to stderr as a unicode QR using ANSI
// colours and half-block characters. Dark pixel → filled, light → empty.
//
// Two pixel rows are combined into one terminal line using ▄ (U+2584) so the
// output is square on typical monospace fonts.
func printQRImage(img image.Image) {
	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	const (
		margin  = 2
		reset   = "\x1b[0m"
		bgWhite = "\x1b[47m"
		bgBlack = "\x1b[40m"
		fgBlack = "\x1b[30m"
	)

	isDark := func(x, y int) bool {
		if x < 0 || y < 0 || x >= w || y >= h {
			return false
		}
		r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
		return (r+g+b)/3 < 0x8000
	}

	// Top margin.
	for i := 0; i < margin; i++ {
		fmt.Fprint(os.Stderr, bgWhite)
		for c := 0; c < w+margin*2; c++ {
			fmt.Fprint(os.Stderr, "  ")
		}
		fmt.Fprintln(os.Stderr, reset)
	}

	for row := 0; row < h; row += 2 {
		// Left margin.
		fmt.Fprint(os.Stderr, bgWhite+fgBlack)
		for m := 0; m < margin; m++ {
			fmt.Fprint(os.Stderr, "  ")
		}

		for col := 0; col < w; col++ {
			top := isDark(col, row)
			bot := isDark(col, row+1)
			switch {
			case top && bot:
				// Both rows dark: solid filled block.
				fmt.Fprint(os.Stderr, bgBlack+fgBlack+"▄"+bgWhite+fgBlack)
			case top && !bot:
				// Upper half dark (use ▀ — upper half block) on white bg.
				fmt.Fprint(os.Stderr, bgWhite+fgBlack+"▀")
			case !top && bot:
				// Lower half dark (▄) on white bg.
				fmt.Fprint(os.Stderr, bgWhite+fgBlack+"▄")
			default:
				// Both light: space.
				fmt.Fprint(os.Stderr, bgWhite+"  ")
			}
		}

		// Right margin.
		fmt.Fprint(os.Stderr, bgWhite+fgBlack)
		for m := 0; m < margin; m++ {
			fmt.Fprint(os.Stderr, "  ")
		}
		fmt.Fprintln(os.Stderr, reset)
	}

	// Bottom margin.
	for i := 0; i < margin; i++ {
		fmt.Fprint(os.Stderr, bgWhite)
		for c := 0; c < w+margin*2; c++ {
			fmt.Fprint(os.Stderr, "  ")
		}
		fmt.Fprintln(os.Stderr, reset)
	}
}
