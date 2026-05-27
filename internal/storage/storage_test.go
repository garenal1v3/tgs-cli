package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewSession_CreatesDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "session.db")

	s, err := NewSession(dbPath)
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	defer func() { _ = s.Close() }()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected db file to be created")
	}
}

func TestSession_StoreAndLoad(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSession(filepath.Join(dir, "session.db"))
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	defer func() { _ = s.Close() }()

	ctx := context.Background()
	data := []byte("test-session-data")

	if err := s.StoreSession(ctx, data); err != nil {
		t.Fatalf("StoreSession() error: %v", err)
	}

	got, err := s.LoadSession(ctx)
	if err != nil {
		t.Fatalf("LoadSession() error: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("LoadSession() = %q, want %q", got, data)
	}
}

func TestSession_LoadEmpty(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSession(filepath.Join(dir, "session.db"))
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	defer func() { _ = s.Close() }()

	ctx := context.Background()
	got, err := s.LoadSession(ctx)
	if err != nil {
		t.Fatalf("LoadSession() error: %v", err)
	}

	if got != nil {
		t.Errorf("LoadSession() on empty db = %q, want nil", got)
	}
}

func TestSession_StoreAndLoadMeta(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSession(filepath.Join(dir, "session.db"))
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	defer func() { _ = s.Close() }()

	if err := s.StoreMeta("phone", []byte("+1234567890")); err != nil {
		t.Fatalf("StoreMeta() error: %v", err)
	}

	got, err := s.LoadMeta("phone")
	if err != nil {
		t.Fatalf("LoadMeta() error: %v", err)
	}

	if string(got) != "+1234567890" {
		t.Errorf("LoadMeta(\"phone\") = %q, want \"+1234567890\"", got)
	}
}

func TestSession_LoadAllMeta(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSession(filepath.Join(dir, "session.db"))
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	defer func() { _ = s.Close() }()

	if err := s.StoreMeta("phone", []byte("+1234567890")); err != nil {
		t.Fatalf("StoreMeta(phone) error: %v", err)
	}
	if err := s.StoreMeta("username", []byte("testuser")); err != nil {
		t.Fatalf("StoreMeta(username) error: %v", err)
	}

	meta, err := s.LoadAllMeta()
	if err != nil {
		t.Fatalf("LoadAllMeta() error: %v", err)
	}

	if meta["phone"] != "+1234567890" {
		t.Errorf("meta[phone] = %q, want \"+1234567890\"", meta["phone"])
	}
	if meta["username"] != "testuser" {
		t.Errorf("meta[username] = %q, want \"testuser\"", meta["username"])
	}
}

func TestSession_LoadMeta_Empty(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSession(filepath.Join(dir, "session.db"))
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	defer func() { _ = s.Close() }()

	got, err := s.LoadMeta("nonexistent")
	if err != nil {
		t.Fatalf("LoadMeta() error: %v", err)
	}
	if got != nil {
		t.Errorf("LoadMeta(nonexistent) = %q, want nil", got)
	}
}
