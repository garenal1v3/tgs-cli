package sources

import (
	"testing"
)

func TestFoldersCmd_FlagParsing(t *testing.T) {
	cmd := newFoldersCmd()
	if cmd.Use != "folders" {
		t.Errorf("Use = %q", cmd.Use)
	}
	if cmd.Flag("archived") == nil {
		t.Error("missing --archived")
	}
	if cmd.Flag("max-wait") == nil {
		t.Error("missing --max-wait")
	}
	if cmd.Flag("no-cache") == nil {
		t.Error("missing --no-cache")
	}
	if cmd.Flag("profile") == nil {
		t.Error("missing --profile")
	}
}
