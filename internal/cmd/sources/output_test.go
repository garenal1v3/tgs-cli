package sources

import (
	"bytes"
	"strings"
	"testing"

	sourcessvc "github.com/searchtgcli/tgs/internal/sources"
)

func TestWriteSource_TextHasKeyLines(t *testing.T) {
	src := &sourcessvc.Source{
		ID: -1001234567890, Type: "channel", Title: "Durov",
		Username: "durov", Access: "public",
		UnreadCount: 0,
	}
	var buf bytes.Buffer
	if err := writeSource(&buf, "text", src); err != nil {
		t.Fatalf("writeSource: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"id:", "type:", "title:", "username:", "Durov", "durov", "channel"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s\n---", want, out)
		}
	}
}

func TestWriteSourceList_JSON(t *testing.T) {
	result := &sourcessvc.ListResult{
		Sources: []sourcessvc.Source{{ID: 1, Type: "user", Title: "Alice"}},
		Total:   1,
	}
	var buf bytes.Buffer
	if err := writeSourceList(&buf, "json", result); err != nil {
		t.Fatalf("writeSourceList: %v", err)
	}
	if !strings.Contains(buf.String(), `"id":1`) {
		t.Errorf("expected id:1 in JSON, got: %s", buf.String())
	}
}
