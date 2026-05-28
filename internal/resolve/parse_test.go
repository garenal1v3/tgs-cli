package resolve

import (
	"strings"
	"testing"
)

func TestParseInput_Username(t *testing.T) {
	tests := []struct {
		input   string
		wantVal string
	}{
		{"@durov", "durov"},
		{"durov", "durov"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pi, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pi.Type != InputUsername {
				t.Errorf("Type = %d, want InputUsername (%d)", pi.Type, InputUsername)
			}
			if pi.Value != tt.wantVal {
				t.Errorf("Value = %q, want %q", pi.Value, tt.wantVal)
			}
			if pi.ID != 0 {
				t.Errorf("ID = %d, want 0", pi.ID)
			}
		})
	}
}

func TestParseInput_NumericID(t *testing.T) {
	pi, err := ParseInput("123456789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pi.Type != InputID {
		t.Errorf("Type = %d, want InputID (%d)", pi.Type, InputID)
	}
	if pi.ID != 123456789 {
		t.Errorf("ID = %d, want 123456789", pi.ID)
	}
	if pi.Value != "" {
		t.Errorf("Value = %q, want empty", pi.Value)
	}
}

func TestParseInput_NegativeID(t *testing.T) {
	pi, err := ParseInput("-1001234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pi.Type != InputID {
		t.Errorf("Type = %d, want InputID (%d)", pi.Type, InputID)
	}
	if pi.ID != -1001234567890 {
		t.Errorf("ID = %d, want -1001234567890", pi.ID)
	}
	if pi.Value != "" {
		t.Errorf("Value = %q, want empty", pi.Value)
	}
}

func TestParseInput_IDPrefix(t *testing.T) {
	// id:<n> is the cobra-friendly alternative to bare negative numbers like
	// "-1001234567890" which pflag interprets as flags.
	tests := []struct {
		input string
		want  int64
	}{
		{"id:42", 42},
		{"id:-1001234567890", -1001234567890},
		{"id:-123456", -123456},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pi, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pi.Type != InputID {
				t.Errorf("Type = %d, want InputID", pi.Type)
			}
			if pi.ID != tt.want {
				t.Errorf("ID = %d, want %d", pi.ID, tt.want)
			}
		})
	}
}

func TestParseInput_IDPrefix_Invalid(t *testing.T) {
	for _, input := range []string{"id:", "id:abc", "id:1.5", "id: 5 "} {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseInput(input); err == nil {
				t.Fatalf("expected error for %q, got nil", input)
			}
		})
	}
}

func TestParseInput_Phone(t *testing.T) {
	pi, err := ParseInput("+79001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pi.Type != InputPhone {
		t.Errorf("Type = %d, want InputPhone (%d)", pi.Type, InputPhone)
	}
	if pi.Value != "79001234567" {
		t.Errorf("Value = %q, want %q", pi.Value, "79001234567")
	}
	if pi.ID != 0 {
		t.Errorf("ID = %d, want 0", pi.ID)
	}
}

func TestParseInput_TmeLink(t *testing.T) {
	tests := []struct {
		input    string
		wantType InputType
		wantVal  string
	}{
		{"t.me/durov", InputUsername, "durov"},
		{"https://t.me/durov", InputUsername, "durov"},
		{"https://t.me/+abc123invite", InputInvite, "abc123invite"},
		{"t.me/+abc123invite", InputInvite, "abc123invite"},
		{"telegram.me/durov", InputUsername, "durov"},
		{"https://telegram.me/durov", InputUsername, "durov"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pi, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pi.Type != tt.wantType {
				t.Errorf("Type = %d, want %d", pi.Type, tt.wantType)
			}
			if pi.Value != tt.wantVal {
				t.Errorf("Value = %q, want %q", pi.Value, tt.wantVal)
			}
			if pi.ID != 0 {
				t.Errorf("ID = %d, want 0", pi.ID)
			}
		})
	}
}

// TestParseInput_InvalidUsername guards the username pre-validation: clearly
// broken inputs must fail with a local error instead of being shipped to
// Telegram only to come back as USERNAME_INVALID. Covers garbage like @@@,
// punctuation, leading digits.
func TestParseInput_InvalidUsername(t *testing.T) {
	cases := []string{
		"@@@",
		"@@durov",
		"@!hi",
		"@hi.there",
		"@1abc", // starts with a digit
		"@a b",  // contains space
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			_, err := ParseInput(in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", in)
			}
			if !strings.Contains(err.Error(), "username") {
				t.Errorf("error = %q, want mention of 'username'", err)
			}
		})
	}
}

// TestParseInput_ShortLegacyUsername guards a regression: Telegram still has
// short legacy usernames (e.g. /test, 4-char handles in some bot APIs); we
// shouldn't reject them locally — let the server be the source of truth on
// the lower bound.
func TestParseInput_ShortLegacyUsername(t *testing.T) {
	for _, in := range []string{"@abcd", "abcd", "ab"} {
		t.Run(in, func(t *testing.T) {
			pi, err := ParseInput(in)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", in, err)
			}
			if pi.Type != InputUsername {
				t.Errorf("Type = %d, want InputUsername", pi.Type)
			}
		})
	}
}

func TestParseInput_Empty(t *testing.T) {
	tests := []string{"", "   "}
	for _, input := range tests {
		t.Run("empty", func(t *testing.T) {
			_, err := ParseInput(input)
			if err == nil {
				t.Fatal("expected error for empty input, got nil")
			}
		})
	}
}
