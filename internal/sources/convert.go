package sources

import (
	"github.com/gotd/td/tg"
)

// botAPIChannelID converts a Telegram channel ID to Bot API negative form
// (e.g. 1234567890 -> -1001234567890).
func botAPIChannelID(id int64) int64 { return -1000000000000 - id }

// activeUsername returns the user-visible handle for a peer. It prefers the
// legacy single Username field; if that is empty it falls back to the first
// active entry from the collectible-usernames array (Telegram's newer
// "additional usernames" model — see tg.Username with Active=true). Without
// this fallback, well-known peers like @durov or @money — whose primary
// handle now lives in Usernames[].Username with Active=true while the legacy
// Username field is empty — silently appear as access=private, username="".
func activeUsername(legacy string, list []tg.Username) string {
	if legacy != "" {
		return legacy
	}
	for _, u := range list {
		if u.Active && u.Username != "" {
			return u.Username
		}
	}
	return ""
}

// chatToSource converts a tg.ChatClass into a partial Source (without
// LastMessage/UnreadCount/Stats/full info). Returns zero Source on unknown types.
func chatToSource(c tg.ChatClass) Source {
	switch v := c.(type) {
	case *tg.Chat:
		s := Source{
			ID:           -v.ID,
			Type:         "group",
			Title:        v.Title,
			MembersCount: v.ParticipantsCount,
		}
		return s

	case *tg.Channel:
		s := Source{
			ID:    botAPIChannelID(v.ID),
			Title: v.Title,
		}
		if v.Broadcast {
			s.Type = "channel"
		} else {
			s.Type = "supergroup"
		}
		legacy, _ := v.GetUsername()
		extra, _ := v.GetUsernames()
		if un := activeUsername(legacy, extra); un != "" {
			s.Username = un
			s.Access = "public"
		} else {
			s.Access = "private"
		}
		if mc, ok := v.GetParticipantsCount(); ok {
			s.MembersCount = mc
		}
		s.Verified = v.Verified
		s.Scam = v.Scam
		s.Fake = v.Fake
		s.Restricted = v.Restricted
		if v.Restricted && len(v.RestrictionReason) > 0 {
			s.RestrictedReason = v.RestrictionReason[0].Reason
		}
		s.HasTopics = v.Forum
		s.Gigagroup = v.Gigagroup
		return s
	}
	return Source{}
}

// userToSource converts a tg.User into a Source. selfID is the authenticated
// user's ID (used to mark Saved Messages).
func userToSource(u *tg.User, selfID int64) Source {
	s := Source{ID: u.ID}
	if u.Bot {
		s.Type = "bot"
	} else {
		s.Type = "user"
	}
	if fn, ok := u.GetFirstName(); ok {
		s.FirstName = fn
	}
	if ln, ok := u.GetLastName(); ok {
		s.LastName = ln
	}
	legacy, _ := u.GetUsername()
	extra, _ := u.GetUsernames()
	if un := activeUsername(legacy, extra); un != "" {
		s.Username = un
	}
	if ph, ok := u.GetPhone(); ok {
		s.Phone = ph
	}
	s.Verified = u.Verified
	s.Scam = u.Scam
	s.Fake = u.Fake
	s.Restricted = u.Restricted
	if u.Restricted && len(u.RestrictionReason) > 0 {
		s.RestrictedReason = u.RestrictionReason[0].Reason
	}
	s.Deleted = u.Deleted
	if selfID != 0 && u.ID == selfID {
		s.Saved = true
		s.Title = "Saved Messages"
	}
	return s
}
