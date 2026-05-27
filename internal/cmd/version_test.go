package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVersionCmd_JSON(t *testing.T) {
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version", "--output", "json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v map[string]string
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	for _, key := range []string{"version", "commit", "date"} {
		if _, ok := v[key]; !ok {
			t.Errorf("missing key %q in version output", key)
		}
	}
}

func TestVersionCmd_Text(t *testing.T) {
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version", "--output", "text"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Error("expected non-empty text output")
	}
}
