package sources

// Source is the unified representation of a Telegram chat/channel/group/user
// as exposed by `tgs sources list` and `tgs sources inspect`.
type Source struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"` // "channel" | "supergroup" | "group" | "user" | "bot"
	Title    string `json:"title,omitempty"`
	Username string `json:"username,omitempty"`
	Access   string `json:"access,omitempty"` // "public" | "private"

	MembersCount int   `json:"members_count,omitempty"`
	HasComments  bool  `json:"has_comments,omitempty"`
	LinkedChatID int64 `json:"linked_chat_id,omitempty"`

	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`

	Verified         bool   `json:"verified,omitempty"`
	Scam             bool   `json:"scam,omitempty"`
	Fake             bool   `json:"fake,omitempty"`
	Restricted       bool   `json:"restricted,omitempty"`
	RestrictedReason string `json:"restricted_reason,omitempty"`
	Archived         bool   `json:"archived,omitempty"`
	Pinned           bool   `json:"pinned,omitempty"`
	Saved            bool   `json:"saved,omitempty"`
	Deleted          bool   `json:"deleted,omitempty"`
	HasTopics        bool   `json:"has_topics,omitempty"`
	Gigagroup        bool   `json:"gigagroup,omitempty"`

	// UnreadCount is always emitted (no omitempty): 0 means "all read", and the
	// field is always populated from the dialog. Omitting it would conflate "all
	// read" with "unknown".
	UnreadCount int          `json:"unread_count"`
	LastMessage *LastMessage `json:"last_message,omitempty"`

	Subscribed   *bool  `json:"subscribed,omitempty"`
	Description  string `json:"description,omitempty"`
	CreationDate string `json:"creation_date,omitempty"`
	InviteLink   string `json:"invite_link,omitempty"`

	Stats      *Stats `json:"stats,omitempty"`
	StatsError string `json:"stats_error,omitempty"`
}

// LastMessage is the most recent message in a dialog.
type LastMessage struct {
	ID   int    `json:"id"`
	Date string `json:"date"` // RFC3339 UTC
}

// FirstMessage is the oldest message in the source.
type FirstMessage struct {
	ID   int    `json:"id"`
	Date string `json:"date"` // RFC3339 UTC
}

// Stats holds the expensive per-source metrics.
type Stats struct {
	TotalMessages int           `json:"total_messages"`
	Messages24h   int           `json:"messages_24h"`
	FirstMessage  *FirstMessage `json:"first_message,omitempty"`
}

// ListRequest is the input for Service.List.
type ListRequest struct {
	Types     []string // filter: "channel" | "supergroup" | "group" | "user" | "bot"; empty = all
	WithStats bool
	Limit     int    // 0 = no cap (fetch all)
	Cursor    string // empty = start
	Archived  bool
}

// ListResult is the output of Service.List.
//
// Returned is len(Sources) — the count after --type filtering and --limit
// trimming. Total is the unfiltered server-reported count for the folder(s)
// walked. The pair lets scripts distinguish "no results because the filter
// matched zero dialogs" (Returned=0 with non-zero Total) from "no results
// because the account is empty" (Total=0) without separately re-reading
// the array length.
type ListResult struct {
	Sources  []Source `json:"sources"`
	Total    int      `json:"total"`
	Returned int      `json:"returned"`
	Cursor   string   `json:"cursor,omitempty"`
}

// InspectRequest is the input for Service.Inspect.
type InspectRequest struct {
	Ref     string // "@username" | "+phone" | numeric ID | "-" | "@me"
	NoStats bool
}

// CachedPayload is the on-disk representation of expensive metrics for a source.
//
// Several fields mirror Source (MembersCount, Description, LinkedChatID,
// HasComments, InviteLink, CreationDate). On cache hit, the service merges
// these onto a fresh Source built from getDialogs; see Service.applyPayload
// (added in a later task). When adding a new expensive field to Source, also
// add it here AND extend applyPayload — otherwise the field will silently be
// zero on cache hits.
type CachedPayload struct {
	MembersCount int    `json:"members_count,omitempty"`
	Description  string `json:"description,omitempty"`
	LinkedChatID int64  `json:"linked_chat_id,omitempty"`
	HasComments  bool   `json:"has_comments,omitempty"`
	InviteLink   string `json:"invite_link,omitempty"`
	CreationDate string `json:"creation_date,omitempty"`
	Stats        *Stats `json:"stats,omitempty"`
}
