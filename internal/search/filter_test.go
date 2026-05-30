package search

import (
	"fmt"
	"testing"

	"github.com/gotd/td/tg"
)

func TestParseFilter_ValidFilters(t *testing.T) {
	expected := map[string]string{
		"photo":       "*tg.InputMessagesFilterPhotos",
		"video":       "*tg.InputMessagesFilterVideo",
		"photo-video": "*tg.InputMessagesFilterPhotoVideo",
		"document":    "*tg.InputMessagesFilterDocument",
		"url":         "*tg.InputMessagesFilterURL",
		"gif":         "*tg.InputMessagesFilterGif",
		"voice":       "*tg.InputMessagesFilterVoice",
		"music":       "*tg.InputMessagesFilterMusic",
		"round-video": "*tg.InputMessagesFilterRoundVideo",
		"geo":         "*tg.InputMessagesFilterGeo",
		"contact":     "*tg.InputMessagesFilterContacts",
		"pinned":      "*tg.InputMessagesFilterPinned",
		"mention":     "*tg.InputMessagesFilterMyMentions",
		"phone-call":  "*tg.InputMessagesFilterPhoneCalls",
		"chat-photo":  "*tg.InputMessagesFilterChatPhotos",
	}

	for name, wantType := range expected {
		t.Run(name, func(t *testing.T) {
			f, err := ParseFilter(name)
			if err != nil {
				t.Fatalf("ParseFilter(%q) returned error: %v", name, err)
			}
			gotType := fmt.Sprintf("%T", f)
			if gotType != wantType {
				t.Errorf("ParseFilter(%q) = %s, want %s", name, gotType, wantType)
			}
		})
	}
}

func TestParseFilter_Empty(t *testing.T) {
	f, err := ParseFilter("")
	if err != nil {
		t.Fatalf("ParseFilter(\"\") returned error: %v", err)
	}
	if _, ok := f.(*tg.InputMessagesFilterEmpty); !ok {
		t.Errorf("ParseFilter(\"\") = %T, want *tg.InputMessagesFilterEmpty", f)
	}
}

func TestParseFilter_Invalid(t *testing.T) {
	_, err := ParseFilter("invalid")
	if err == nil {
		t.Fatal("ParseFilter(\"invalid\") expected error, got nil")
	}
}

func TestAllFilterNames(t *testing.T) {
	names := AllFilterNames()
	if len(names) != 15 {
		t.Fatalf("AllFilterNames() returned %d names, want 15", len(names))
	}
	// Verify sorted order.
	for i := 1; i < len(names); i++ {
		if names[i] <= names[i-1] {
			t.Errorf("AllFilterNames() not sorted: %q <= %q at index %d", names[i], names[i-1], i)
		}
	}
}
