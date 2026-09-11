package session

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/squawk-dev/squawk/internal/model"
)

func TestCreateAndResolveActiveSession(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	sess, err := store.CreateSession("emulator-5554", 300)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	if sess.LogLines != 300 {
		t.Fatalf("expected LogLines 300, got %d", sess.LogLines)
	}

	current, err := store.CurrentSession()
	if err != nil {
		t.Fatalf("CurrentSession: %v", err)
	}
	if current.ID != sess.ID {
		t.Fatalf("expected active session %s, got %s", sess.ID, current.ID)
	}
	if current.Device != "emulator-5554" {
		t.Fatalf("unexpected device: %q", current.Device)
	}
}

func TestCurrentSessionMissingPointer(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := store.CurrentSession(); !errors.Is(err, ErrNoActiveSession) {
		t.Fatalf("expected ErrNoActiveSession, got %v", err)
	}
}

func TestCurrentSessionPointerToMissingSession(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := os.MkdirAll(store.RootDir(), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(store.PointerFile(), []byte("19990101-000000\n"), 0o644); err != nil {
		t.Fatalf("write pointer: %v", err)
	}
	if _, err := store.CurrentSession(); !errors.Is(err, ErrNoActiveSession) {
		t.Fatalf("expected ErrNoActiveSession, got %v", err)
	}
}

func TestCurrentSessionEmptyPointer(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := os.MkdirAll(store.RootDir(), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(store.PointerFile(), []byte("   \n"), 0o644); err != nil {
		t.Fatalf("write pointer: %v", err)
	}
	if _, err := store.CurrentSession(); !errors.Is(err, ErrNoActiveSession) {
		t.Fatalf("expected ErrNoActiveSession, got %v", err)
	}
}

func TestSquawkDirectories(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	sess, err := store.CreateSession("emulator-5554", 200)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	want := filepath.Join(root, "sessions", sess.ID, "squawks", "001")
	if got := store.SquawkDir(sess.ID, 1); got != want {
		t.Fatalf("SquawkDir = %q, want %q", got, want)
	}
}

func TestNextSquawkID(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	sess, err := store.CreateSession("emulator-5554", 200)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if got := NextSquawkID(sess); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
	sess.Squawks = append(sess.Squawks, model.Squawk{ID: 1})
	if got := NextSquawkID(sess); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestDefaultRootDirIsDotSquawk(t *testing.T) {
	store, err := NewStore("")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if filepath.Base(store.RootDir()) != DefaultRootName {
		t.Fatalf("expected default root .squawk, got %q", store.RootDir())
	}
}
