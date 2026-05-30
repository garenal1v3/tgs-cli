package sources

import "testing"

func TestCursor_RoundTrip(t *testing.T) {
	c := &Cursor{
		OffsetDate:           1700000000,
		OffsetID:             4242,
		OffsetPeerType:       "channel",
		OffsetPeerID:         1234567890,
		OffsetPeerAccessHash: -987654321,
	}
	s := c.Encode()
	if s == "" {
		t.Fatal("expected non-empty cursor string")
	}
	got, err := DecodeCursor(s)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if *got != *c {
		t.Errorf("round trip mismatch:\n got  = %+v\n want = %+v", got, c)
	}
}

func TestDecodeCursor_Empty(t *testing.T) {
	got, err := DecodeCursor("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil cursor for empty input, got %+v", got)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	if _, err := DecodeCursor("!!!not-base64!!!"); err == nil {
		t.Error("expected error for invalid input")
	}
}
