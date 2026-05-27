package search

// SearchResult is the top-level response for message search queries.
type SearchResult struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
	Cursor   string    `json:"cursor,omitempty"`
}

// Message represents a single Telegram message in search results.
type Message struct {
	ID           int        `json:"id"`
	Chat         ChatInfo   `json:"chat"`
	From         *UserInfo  `json:"from,omitempty"`
	Date         string     `json:"date"`
	Text         string     `json:"text"`
	Media        *MediaInfo `json:"media,omitempty"`
	ReplyToMsgID int        `json:"reply_to_msg_id,omitempty"`
	TopicID      int        `json:"topic_id,omitempty"`
	Views        int        `json:"views,omitempty"`
	Forwards     int        `json:"forwards,omitempty"`
}

// ChatInfo describes the chat a message belongs to.
type ChatInfo struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	Username string `json:"username,omitempty"`
}

// UserInfo describes the sender of a message.
type UserInfo struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// MediaInfo describes media attached to a message.
type MediaInfo struct {
	Type     string `json:"type"`
	FileSize int64  `json:"file_size,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileName string `json:"file_name,omitempty"`
	Duration int    `json:"duration,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// CountersResult is the response for search counters queries.
type CountersResult struct {
	Counters []CounterEntry `json:"counters"`
}

// CounterEntry represents a count for a specific message filter type.
type CounterEntry struct {
	Filter string `json:"filter"`
	Count  int    `json:"count"`
}

// CalendarResult is the response for search calendar queries.
type CalendarResult struct {
	Periods []CalendarPeriod `json:"periods"`
	Total   int              `json:"total"`
}

// CalendarPeriod represents message activity within a date range.
type CalendarPeriod struct {
	Date     string `json:"date"`
	Count    int    `json:"count"`
	MinMsgID int    `json:"min_msg_id"`
	MaxMsgID int    `json:"max_msg_id"`
}

// ErrorResult is the JSON error response wrapper.
type ErrorResult struct {
	Error ErrorInfo `json:"error"`
}

// ErrorInfo contains structured error details.
type ErrorInfo struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after,omitempty"`
}
