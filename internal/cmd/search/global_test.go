package search

import "testing"

func TestGlobalCmd_FolderFlag(t *testing.T) {
	cmd := newGlobalCmd()
	if cmd.Flag("folder") == nil {
		t.Error("missing --folder")
	}
	if cmd.Flag("archived") == nil {
		t.Error("missing --archived")
	}
}
