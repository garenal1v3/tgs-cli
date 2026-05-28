package sources

import "github.com/gotd/td/tg"

// Folder is a Telegram dialog filter (a.k.a. "folder" in the UI) rendered for
// display: the filter's metadata plus its resolved contents after applying
// pinned/include/exclude rules and auto-filters to the user's dialog list.
//
// Default folder ("All chats") is excluded — Folder values represent only
// user-defined folders (DialogFilter) and shared chatlists (DialogFilterChatlist).
type Folder struct {
	ID           int      `json:"id"`
	Kind         string   `json:"kind"` // "custom" | "chatlist"
	Title        string   `json:"title"`
	Emoticon     string   `json:"emoticon,omitempty"`
	HasMyInvites bool     `json:"has_my_invites,omitempty"` // chatlist-only
	ChatsCount   int      `json:"chats_count"`
	Chats        []Source `json:"chats"`
}

// FoldersRequest is the input for Service.Folders.
type FoldersRequest struct {
	Archived bool // include archived dialogs (folder 1) in the source pool
}

// FoldersResult is the output of Service.Folders.
type FoldersResult struct {
	Folders []Folder `json:"folders"`
	Total   int      `json:"total"`
}

// ResolvedFolder is the result of Service.ResolveFolder. Folder is the same
// shape the discovery command emits; Peers gives search commands the
// InputPeers they need for fan-out searches.
type ResolvedFolder struct {
	Folder Folder
	Peers  []tg.InputPeerClass
}
