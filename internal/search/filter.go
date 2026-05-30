package search

import (
	"fmt"
	"sort"

	"github.com/gotd/td/tg"
)

var filterMap = map[string]func() tg.MessagesFilterClass{
	"photo":       func() tg.MessagesFilterClass { return &tg.InputMessagesFilterPhotos{} },
	"video":       func() tg.MessagesFilterClass { return &tg.InputMessagesFilterVideo{} },
	"photo-video": func() tg.MessagesFilterClass { return &tg.InputMessagesFilterPhotoVideo{} },
	"document":    func() tg.MessagesFilterClass { return &tg.InputMessagesFilterDocument{} },
	"url":         func() tg.MessagesFilterClass { return &tg.InputMessagesFilterURL{} },
	"gif":         func() tg.MessagesFilterClass { return &tg.InputMessagesFilterGif{} },
	"voice":       func() tg.MessagesFilterClass { return &tg.InputMessagesFilterVoice{} },
	"music":       func() tg.MessagesFilterClass { return &tg.InputMessagesFilterMusic{} },
	"round-video": func() tg.MessagesFilterClass { return &tg.InputMessagesFilterRoundVideo{} },
	"geo":         func() tg.MessagesFilterClass { return &tg.InputMessagesFilterGeo{} },
	"contact":     func() tg.MessagesFilterClass { return &tg.InputMessagesFilterContacts{} },
	"pinned":      func() tg.MessagesFilterClass { return &tg.InputMessagesFilterPinned{} },
	"mention":     func() tg.MessagesFilterClass { return &tg.InputMessagesFilterMyMentions{} },
	"phone-call":  func() tg.MessagesFilterClass { return &tg.InputMessagesFilterPhoneCalls{} },
	"chat-photo":  func() tg.MessagesFilterClass { return &tg.InputMessagesFilterChatPhotos{} },
}

// ParseFilter maps a string filter name to a gotd/td MessagesFilterClass.
// An empty string returns InputMessagesFilterEmpty. An unknown name returns an error.
func ParseFilter(name string) (tg.MessagesFilterClass, error) {
	if name == "" {
		return &tg.InputMessagesFilterEmpty{}, nil
	}
	fn, ok := filterMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown filter %q, valid filters: %v", name, AllFilterNames())
	}
	return fn(), nil
}

// AllFilterNames returns a sorted slice of all valid filter names.
func AllFilterNames() []string {
	names := make([]string, 0, len(filterMap))
	for k := range filterMap {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
