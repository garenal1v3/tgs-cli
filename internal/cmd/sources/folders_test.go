package sources

import (
	"bytes"
	"strings"
	"testing"
)

func TestFoldersCmd_FlagParsing(t *testing.T) {
	cmd := newFoldersCmd()
	if cmd.Use != "folders" {
		t.Errorf("Use = %q", cmd.Use)
	}
	// --archived was removed: a folder view always includes archived chats,
	// so the flag would be a no-op (see feedback on folder/archive semantics).
	if cmd.Flag("archived") != nil {
		t.Error("--archived should not exist on 'sources folders'")
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

func TestListCmd_FolderFlag(t *testing.T) {
	cmd := newListCmd()
	if cmd.Flag("folder") == nil {
		t.Error("missing --folder")
	}
}

func TestListCmd_CursorAndFolderMutuallyExclusive(t *testing.T) {
	cmd := newListCmd()
	cmd.SetArgs([]string{"--cursor", "abc", "--folder", "Whatever"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cursor") || !strings.Contains(err.Error(), "folder") {
		t.Errorf("err = %v, want mention of cursor+folder", err)
	}
}
