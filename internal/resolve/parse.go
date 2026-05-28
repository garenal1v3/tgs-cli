package resolve

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

// InputType represents the kind of peer reference parsed from user input.
type InputType int

const (
	InputUsername InputType = iota
	InputID
	InputPhone
	InputInvite
)

// PeerInput holds the parsed result of a raw input string.
type PeerInput struct {
	Type  InputType
	Value string
	ID    int64
}

// ErrEmptyInput is returned when the input string is empty.
var ErrEmptyInput = errors.New("empty input")

// ParseInput determines the type of peer reference from a raw string.
//
// Supported formats:
//   - @durov or durov         → InputUsername
//   - 123456789               → InputID (positive)
//   - -1001234567890          → InputID (negative / supergroup)
//   - +79001234567            → InputPhone (7+ digits after +, no / in string)
//   - t.me/durov              → InputUsername
//   - https://t.me/durov      → InputUsername
//   - t.me/+abc123invite      → InputInvite
//   - telegram.me/...         → same as t.me
func ParseInput(input string) (PeerInput, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return PeerInput{}, ErrEmptyInput
	}

	// Try parsing as a Telegram link first.
	if pi, ok := parseTelegramLink(input); ok {
		return pi, nil
	}

	// Explicit ID form `id:<n>`: cobra/pflag treats bare "-1001234567890" as
	// a flag, so this is the CLI-friendly way to pass a numeric peer ID
	// without needing the `--` separator.
	if rest, ok := strings.CutPrefix(input, "id:"); ok {
		id, err := strconv.ParseInt(rest, 10, 64)
		if err != nil {
			return PeerInput{}, fmt.Errorf("invalid id:<n> form: %q (want a base-10 integer)", input)
		}
		return PeerInput{Type: InputID, ID: id}, nil
	}

	// Phone: starts with +, followed by 7+ digits, no slash in string.
	if strings.HasPrefix(input, "+") && !strings.Contains(input, "/") {
		digits := input[1:]
		if len(digits) >= 7 && allDigits(digits) {
			return PeerInput{Type: InputPhone, Value: digits}, nil
		}
	}

	// Numeric ID (positive or negative).
	if id, err := strconv.ParseInt(input, 10, 64); err == nil {
		return PeerInput{Type: InputID, ID: id}, nil
	}

	// Username: strip leading @.
	username := strings.TrimPrefix(input, "@")
	if username == "" {
		return PeerInput{}, fmt.Errorf("invalid input: %q", input)
	}
	if !isValidUsername(username) {
		return PeerInput{}, fmt.Errorf("invalid username form: %q (Telegram usernames may contain only letters, digits and underscores, and must start with a letter)", input)
	}
	return PeerInput{Type: InputUsername, Value: username}, nil
}

// isValidUsername mirrors Telegram's username rules just well enough to
// reject obviously broken inputs (@@@, @!hi, @1abc) up front instead of
// letting Telegram return a confusing USERNAME_INVALID code. We intentionally
// do not enforce the 5-character minimum: short legacy usernames exist (and
// the Bot API still accepts them), so leave that boundary to the server.
func isValidUsername(s string) bool {
	if len(s) == 0 || len(s) > 32 {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
		if i == 0 && (r >= '0' && r <= '9') {
			// First char must be a letter or underscore; digits aren't
			// legal as the first character.
			return false
		}
	}
	return true
}

// parseTelegramLink attempts to extract a peer reference from t.me or telegram.me URLs.
func parseTelegramLink(input string) (PeerInput, bool) {
	// Normalise: if there is no scheme, prepend https:// so url.Parse works.
	raw := input
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return PeerInput{}, false
	}

	host := strings.ToLower(u.Hostname())
	if host != "t.me" && host != "telegram.me" {
		return PeerInput{}, false
	}

	// Path should be like /durov or /+invite_hash
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return PeerInput{}, false
	}

	// Invite link: /+<hash>
	if strings.HasPrefix(path, "+") {
		hash := path[1:]
		if hash == "" {
			return PeerInput{}, false
		}
		return PeerInput{Type: InputInvite, Value: hash}, true
	}

	// Otherwise treat as username (first path segment only).
	segments := strings.SplitN(path, "/", 2)
	username := segments[0]
	if username == "" {
		return PeerInput{}, false
	}
	return PeerInput{Type: InputUsername, Value: username}, true
}

// allDigits reports whether s is non-empty and every rune is an ASCII digit.
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
