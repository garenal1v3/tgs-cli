package resolve

import (
	"testing"
)

func TestParseInput_Username(t *testing.T) {
	tests := []struct {
		input    string
		wantVal  string
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
