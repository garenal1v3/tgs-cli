package search

import (
	"encoding/base64"
	"encoding/json"
)

// Cursor represents pagination state for a single-peer search.
type Cursor struct {
	OffsetID   int    `json:"o"`
	OffsetDate int    `json:"d"`
	OffsetRate int    `json:"r,omitempty"`
	OffsetPeer []byte `json:"p,omitempty"`
}

// Encode serialises the cursor as base64 RawURL-encoded JSON.
func (c *Cursor) Encode() string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor decodes a base64 RawURL-encoded JSON cursor.
// An empty string returns (nil, nil).
func DecodeCursor(s string) (*Cursor, error) {
	if s == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// MultiCursor represents pagination state across multiple chats.
type MultiCursor struct {
	Chats []ChatCursor `json:"c"`
}

// ChatCursor represents pagination state for one chat within a multi-cursor.
type ChatCursor struct {
	PeerID   int64 `json:"id"`
	OffsetID int   `json:"o"`
	Done     bool  `json:"done"`
}

// Encode serialises the multi-cursor as base64 RawURL-encoded JSON.
func (mc *MultiCursor) Encode() string {
	b, _ := json.Marshal(mc)
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeMultiCursor decodes a base64 RawURL-encoded JSON multi-cursor.
// An empty string returns (nil, nil).
func DecodeMultiCursor(s string) (*MultiCursor, error) {
	if s == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var mc MultiCursor
	if err := json.Unmarshal(b, &mc); err != nil {
		return nil, err
	}
	return &mc, nil
}
