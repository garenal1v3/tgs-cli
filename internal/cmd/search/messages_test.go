package search

import "testing"

func TestMessagesCmd_FolderFlag(t *testing.T) {
	cmd := newMessagesCmd()
	if cmd.Flag("folder") == nil {
		t.Error("missing --folder")
	}
}
