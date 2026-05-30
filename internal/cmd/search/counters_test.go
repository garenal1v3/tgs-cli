package search

import "testing"

func TestCountersCmd_FolderFlag(t *testing.T) {
	cmd := newCountersCmd()
	if cmd.Flag("folder") == nil {
		t.Error("missing --folder")
	}
}
