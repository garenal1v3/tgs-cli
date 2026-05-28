package search

import "testing"

func TestCalendarCmd_FolderFlag(t *testing.T) {
	cmd := newCalendarCmd()
	if cmd.Flag("folder") == nil {
		t.Error("missing --folder")
	}
}
