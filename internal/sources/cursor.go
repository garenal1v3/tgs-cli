package sources

import (
	"encoding/base64"
	"encoding/json"
)

// Cursor encodes pagination state for messages.getDialogs.
type Cursor struct {
	OffsetDate           int    `json:"d"`
	OffsetID             int    `json:"o"`
	OffsetPeerType       string `json:"pt,omitempty"` // "user" | "chat" | "channel" | ""
	OffsetPeerID         int64  `json:"pi,omitempty"`
	OffsetPeerAccessHash int64  `json:"ph,omitempty"`
}

// Encode serialises the cursor as base64 RawURL-encoded JSON.
func (c *Cursor) Encode() string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor decodes a base64 RawURL-encoded JSON cursor. Empty -> (nil, nil).
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
