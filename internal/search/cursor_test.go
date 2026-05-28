package search

import (
	"testing"
)

func TestCursor_EncodeDecode(t *testing.T) {
	original := &Cursor{
		OffsetID:   42,
		OffsetDate: 1700000000,
	}
	encoded := original.Encode()
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor() error: %v", err)
	}
	if decoded.OffsetID != original.OffsetID {
		t.Errorf("OffsetID = %d, want %d", decoded.OffsetID, original.OffsetID)
	}
	if decoded.OffsetDate != original.OffsetDate {
		t.Errorf("OffsetDate = %d, want %d", decoded.OffsetDate, original.OffsetDate)
	}
}

func TestCursor_GlobalFields(t *testing.T) {
	original := &Cursor{
		OffsetID:   10,
		OffsetDate: 1700000000,
		OffsetRate: 99,
		OffsetPeer: []byte{1, 2, 3},
	}
	encoded := original.Encode()
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor() error: %v", err)
	}
	if decoded.OffsetRate != original.OffsetRate {
		t.Errorf("OffsetRate = %d, want %d", decoded.OffsetRate, original.OffsetRate)
	}
	if len(decoded.OffsetPeer) != len(original.OffsetPeer) {
		t.Fatalf("OffsetPeer length = %d, want %d", len(decoded.OffsetPeer), len(original.OffsetPeer))
	}
	for i, b := range decoded.OffsetPeer {
		if b != original.OffsetPeer[i] {
			t.Errorf("OffsetPeer[%d] = %d, want %d", i, b, original.OffsetPeer[i])
		}
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	_, err := DecodeCursor("!!!not-valid-base64!!!")
	if err == nil {
		t.Fatal("DecodeCursor(garbage) expected error, got nil")
	}
}

func TestDecodeCursor_Empty(t *testing.T) {
	c, err := DecodeCursor("")
	if err != nil {
		t.Fatalf("DecodeCursor(\"\") returned error: %v", err)
	}
	if c != nil {
		t.Errorf("DecodeCursor(\"\") = %+v, want nil", c)
	}
}

func TestMultiCursor_EncodeDecode(t *testing.T) {
	original := &MultiCursor{
		Chats: []ChatCursor{
			{PeerID: 100, OffsetID: 50, Done: false},
			{PeerID: 200, OffsetID: 0, Done: true},
		},
	}
	encoded := original.Encode()
	decoded, err := DecodeMultiCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeMultiCursor() error: %v", err)
	}
	if len(decoded.Chats) != 2 {
		t.Fatalf("Chats length = %d, want 2", len(decoded.Chats))
	}
	for i, chat := range decoded.Chats {
		want := original.Chats[i]
		if chat.PeerID != want.PeerID {
			t.Errorf("Chats[%d].PeerID = %d, want %d", i, chat.PeerID, want.PeerID)
		}
		if chat.OffsetID != want.OffsetID {
			t.Errorf("Chats[%d].OffsetID = %d, want %d", i, chat.OffsetID, want.OffsetID)
		}
		if chat.Done != want.Done {
			t.Errorf("Chats[%d].Done = %v, want %v", i, chat.Done, want.Done)
		}
	}
}
