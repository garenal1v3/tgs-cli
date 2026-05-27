package telegram

import (
	"testing"
)

func TestAPICredentials_Defaults(t *testing.T) {
	t.Setenv("TGS_API_ID", "")
	t.Setenv("TGS_API_HASH", "")

	id, hash := apiCredentials()
	if id != 0 {
		t.Errorf("apiCredentials() id = %d, want 0", id)
	}
	if hash != defaultAPIHash {
		t.Errorf("apiCredentials() hash = %q, want %q", hash, defaultAPIHash)
	}
}

func TestAPICredentials_EnvOverride(t *testing.T) {
	t.Setenv("TGS_API_ID", "99999")
	t.Setenv("TGS_API_HASH", "override-hash")

	id, hash := apiCredentials()
	if id != 99999 {
		t.Errorf("apiCredentials() id = %d, want 99999", id)
	}
	if hash != "override-hash" {
		t.Errorf("apiCredentials() hash = %q, want \"override-hash\"", hash)
	}
}

func TestAPICredentials_InvalidEnvID(t *testing.T) {
	t.Setenv("TGS_API_ID", "not-a-number")
	t.Setenv("TGS_API_HASH", "some-hash")

	id, _ := apiCredentials()
	if id != 0 {
		t.Errorf("apiCredentials() with invalid env id = %d, want 0", id)
	}
}
