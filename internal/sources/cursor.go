package sources

import (
	"encoding/base64"
	"encoding/json"
)

// Cursor encodes pagination state for messages.getDialogs.
//
// Folder distinguishes Telegram's main folder (0) from the archive (1).
// messages.getDialogs returns only one folder at a time; when --archived is
// set the List walks folder 0 to completion and then continues with folder 1.
// The cursor must remember which folder we're on so that pagination across
// process boundaries (multiple CLI invocations) resumes correctly.
type Cursor struct {
	OffsetDate           int    `json:"d"`
	OffsetID             int    `json:"o"`
	OffsetPeerType       string `json:"pt,omitempty"` // "user" | "chat" | "channel" | ""
	OffsetPeerID         int64  `json:"pi,omitempty"`
	OffsetPeerAccessHash int64  `json:"ph,omitempty"`
	Folder               int    `json:"f,omitempty"` // 0 = main (default), 1 = archive
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
